package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

var (
	log     = logrus.New()

	//List taken from
	// https://docs.openshift.com/container-platform/4.2/networking/multiple_networks/configuring-sr-iov.html#supported-devices_configuring-sr-iov
	// 0x17a0:0x9750 added for testing on my laptop (it's an SD
	// card reader which is interesting solely because I'm not
	// generally using it on the host) -dgibson
	supportedPciDevices = map[string]bool{
		"0x8086:0x1521": true,
		"0x8086:0x1520": true,
		"0x8086:0x158b": true,
		"0x8086:0x154c": true,
		"0x15b3:0x1015": true,
		"0x15b3:0x1017": true,
		"0x17a0:0x9750": true,
	}
)

const (
	pciDeviceFile  = "/sys/bus/pci/devices"
	vfioDeviceFile = "/sys/bus/pci/drivers/vfio-pci"

	waitGranule = 100 * time.Millisecond
	waitTimeout = 10 * time.Second
)

type GuestPciPath struct {
	// XXX If we want to support machine types with multiple PCI
	// roots, we'll need to add a machine specific way of
	// describing the correct root here

	// A list of slot/function pairs for each bridge leading to
	// the device, then finally for the device itself
	Path []string
}

type vfioDevInfo struct {
	HostAddress string                `json:"host-address"`
	GuestPath GuestPciPath            `json:"guest-path"`
}
type vfioGroupInfo struct {
	HostGroup string                   `json:"host-group"`
	Devices []vfioDevInfo              `json:"devices"`
}
type vfioInfo []vfioGroupInfo

func main() {
	log.Out = os.Stderr
	var err error

	dname, err := ioutil.TempDir("", "vfiohooklog")
	fname := filepath.Join(dname, "vfiohook.log")
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Warning("Failed to log to file, using default stderr")
	}
	//Set default to Debug for debugging during dev
	log.SetLevel(logrus.DebugLevel)
	log.Debugf("Started VFIO OCI hook")

	//This can be used when the hook needs to be run standalone
	force := flag.Bool("f", false, "Force start the VFIO hook")
	flag.Parse()

	var cid string // Container ID

	if !*force {
		cid, err = parseState()
	} else {
		// Force option doesn't get state, since it's run
		// without stdin wired up to the agent, so we have to
		// look elsewhere for the container ID
		cid, err = guessID()
	}
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Infof("VFIO hook, container %s", cid)

	config, err := loadConfig(cid)
	if err != nil {
		log.Fatal(err)
		return
	}

	info, err := getVfioInfo(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = waitForDevices(info)
	if err != nil {
		log.Fatal(err)
		return
	}

	scanDevices()
}

func parseState() (string, error) {
	var s specs.State

	reader := bufio.NewReader(os.Stdin)
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&s)
	if err != nil {
		return "", err
	}

	log.Debugf("Container state: %v", s)
	return s.ID, nil
}

func guessID() (string, error) {
	cList, err := ioutil.ReadDir("/run/libcontainer")
	if err != nil {
		return "", err
	}

	if len(cList) != 1 {
		return "", fmt.Errorf("Couldn't identify container ID")
	}

	return cList[0].Name(), nil
}

func loadConfig(cid string) (*specs.Spec, error) {
	//For Kata the config.json isn't in the bundle path
	configJsonPath := filepath.Join("/run/libcontainer", cid, "config.json")

	log.Debugf("Reading config.json from:: %s", configJsonPath)

	configFile, err := os.OpenFile(configJsonPath, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(configFile)
	decoder := json.NewDecoder(reader)

	var config specs.Spec

	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func getVfioInfo(config *specs.Spec) (*vfioInfo, error) {
	annotation := config.Annotations["io.katacontainers.x.vfio"]
	log.Debugf("VFIO device annotation: %s", annotation)

	var info vfioInfo

	err := json.Unmarshal([]byte(annotation), &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func waitForPciPath(pciPath GuestPciPath) (string, error) {
	path := pciPath.Path
	var bdf string
	var err error

	// XXX this is x86(pc/q35) specific for now
	b := "0000:00"
	sysPath := fmt.Sprintf("/sys/devices/pci%s", b)

	if len(path) == 0 {
		return "", fmt.Errorf("Empty guest PCI path")
	}

	for {
		df := path[0]
		path = path[1:]
		bdf = fmt.Sprintf("%s:%s", b, df)
		sysPath = filepath.Join(sysPath, bdf)
		log.Debugf("Waiting for %s", sysPath)

		then := time.Now()
		for {
			_, err = os.Stat(sysPath)
			if err == nil {
				break
			}
			if !os.IsNotExist(err) {
				return "", err
			}
			if elapsed := time.Since(then); elapsed > waitTimeout {
				return "", fmt.Errorf("Timeout waiting for %s (%v)", sysPath, elapsed)
			}
			time.Sleep(waitGranule)
		}

		log.Debugf("Found %s after %v", sysPath, time.Since(then))

		if len(path) == 0 {
			// We're there!
			return bdf, nil
		}

		busDir := filepath.Join(sysPath, "pci_bus")
		busList, err := ioutil.ReadDir(busDir)
		if err != nil {
			return "", err
		}

		if len(busList) != 1 {
			return "", fmt.Errorf("Unexpected contents of %s", busDir)
		}

		b = busList[0].Name()
	}
}

func waitForDevices(info *vfioInfo) error {
	for _, group := range *info {
		log.Debugf("waitForDevice: Host Group %s", group.HostGroup)

		for _, dev := range group.Devices {
			guestBDF, err := waitForPciPath(dev.GuestPath)
			if err != nil {
				return err
			}
			log.Infof("Host device %s is guest device %s",
				dev.HostAddress, guestBDF)
		}
	}

	return nil
}

func scanDevices() {
	bdfList, err := ioutil.ReadDir(pciDeviceFile)
	if err != nil {
		log.Errorf("Failed to readdir %s: %s", pciDeviceFile, err)
		return
	}

	for _, bdf := range bdfList {
		log.Debugf("bdf = ", bdf)
		vendorPath := filepath.Join(pciDeviceFile, bdf.Name(), "vendor")
		devicePath := filepath.Join(pciDeviceFile, bdf.Name(), "device")
		vendor, err := ioutil.ReadFile(vendorPath)
		if err != nil {
			log.Errorf("Failed to read %s: %s", vendorPath, err)
			continue
		}
		device, err := ioutil.ReadFile(devicePath)
		if err != nil {
			log.Errorf("Failed to read %s: %s", devicePath, err)
			continue
		}
		vd := fmt.Sprintf("%s:%s", strings.TrimSuffix(string(vendor), "\n"), strings.TrimSuffix(string(device), "\n"))
		log.Debugf("vd = ", vd)

		if supportedPciDevices[vd] {
			err = rebindOne(bdf.Name(), vd)
			if err != nil {
				log.Errorf("%s", err)
				continue
			}
		}
	}

	return
}

func rebindOne(bdf string, vd string) error {
	log.Debugf("Attempting to rebind device %s to vfio", bdf)

	driverPath := filepath.Join(pciDeviceFile, bdf, "driver")
	if _, err := os.Stat(driverPath); err == nil {
		driver, err := os.Readlink(driverPath)
		if err != nil {
			return fmt.Errorf("Failed to read %s: %s", driverPath, err)
		}
		if string(driver) == "vfio-pci" {
			log.Infof("Device %s is already bound to vfio", bdf)
			return nil
		} else {
			log.Infof("Unbinding %s from driver (%s)", bdf, string(driver))
			unbindPath := filepath.Join(pciDeviceFile, bdf, "driver/unbind")
			err = ioutil.WriteFile(unbindPath, []byte(bdf), 0200)
			if err != nil {
				return fmt.Errorf("Failed to unbind driver (%s): %s", string(driver), err)
			}
		}
	}

	newidPath := filepath.Join(vfioDeviceFile, "new_id")
	newid := strings.Replace(vd, ":", " ", 1)
	err := ioutil.WriteFile(newidPath, []byte(newid), 0200)
	if err != nil {
		return fmt.Errorf("Failed to write %s: %s", vfioDeviceFile, err)
	}
	log.Infof("Completed rebinding device %s to vfio", bdf)
	return nil
}
