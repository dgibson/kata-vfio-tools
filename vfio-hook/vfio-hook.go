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
)

func main() {
	log.Out = os.Stderr

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

	_, err = loadConfig(cid)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Wait for devices to be ready
	time.Sleep(10 * time.Second)

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

func loadConfig(cid string) ([]byte, error) {
	//For Kata the config.json isn't in the bundle path
	configJsonPath := filepath.Join("/run/libcontainer", cid, "config.json")

	log.Debugf("Reading config.json from:: %s", configJsonPath)

	jsonData, err := ioutil.ReadFile(configJsonPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config.json: %s", err)
	}

	log.Debugf("Config.json contents: %s", jsonData)
	return jsonData, nil
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
