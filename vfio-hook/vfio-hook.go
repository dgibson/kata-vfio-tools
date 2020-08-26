package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

const (
	pciDeviceFile  = "/sys/bus/pci/devices"
	pciProbeFile   = "/sys/bus/pci/drivers_probe"

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
	GuestBDF string
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

	err = rebindDevices(info)
	if err != nil {
		log.Fatal(err)
		return
	}
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
	for i, group := range *info {
		log.Debugf("waitForDevice: Host Group %s", group.HostGroup)

		for j, dev := range group.Devices {
			guestBDF, err := waitForPciPath(dev.GuestPath)
			if err != nil {
				return err
			}
			log.Infof("Host device %s is guest device %s",
				dev.HostAddress, guestBDF)
			(*info)[i].Devices[j].GuestBDF = guestBDF
		}
	}

	return nil
}

func rebindDevices(info *vfioInfo) error {
	for _, group := range *info {
		for _, dev := range group.Devices {
			err := rebindOne(dev.GuestBDF)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func rebindOne(bdf string) error {
	log.Debugf("Attempting to rebind device %s to vfio", bdf)

	driverPath := filepath.Join(pciDeviceFile, bdf, "driver")
	driver, err := os.Readlink(driverPath)
	if os.IsNotExist(err) {
		// This just means the device isn't bound
		driver, err = "", nil
	}
	if err != nil {
		return fmt.Errorf("Failed to read %s: %s", driverPath, err)
	}
	log.Debugf("Device %s previously bound to '%s'", bdf, driver)

	if string(driver) == "vfio-pci" {
		log.Infof("Device %s is already bound to vfio", bdf)
		return nil
	}

	overridePath := filepath.Join(pciDeviceFile, bdf, "driver_override")
	err = ioutil.WriteFile(overridePath, []byte("vfio-pci"), 0200)
	if err != nil {
		return fmt.Errorf("Failed to override driver via %s", overridePath)
	}

	if driver != "" {
		log.Debugf("Unbinding %s from driver (%s)", bdf, string(driver))
		unbindPath := filepath.Join(pciDeviceFile, bdf, "driver/unbind")
		err = ioutil.WriteFile(unbindPath, []byte(bdf), 0200)
		if err != nil {
			return fmt.Errorf("Failed to unbind driver (%s): %s", string(driver), err)
		}
	}

	err = ioutil.WriteFile(pciProbeFile, []byte(bdf), 0200)
	if err != nil {
		return fmt.Errorf("Failed to reprobe %s: %s", bdf, err)
	}
	log.Infof("Device %s rebound to vfio", bdf)

	return nil
}
