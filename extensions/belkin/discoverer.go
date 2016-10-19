package belkin

import (
	"fmt"
	"strings"

	belkinExt "github.com/go-home-iot/belkin"
	"github.com/markdaws/gohome"
	"github.com/markdaws/gohome/log"
	"github.com/markdaws/gohome/zone"
)

type discoverer struct {
	System *gohome.System
}

func (d *discoverer) Devices(sys *gohome.System, modelNumber string) ([]gohome.Device, error) {

	log.V("scanning belkin")
	var scanType belkinExt.DeviceType
	switch modelNumber {
	case "f7c043fc":
		scanType = belkinExt.DTMaker
	case "f7c029v2":
		scanType = belkinExt.DTInsight
	default:
		return nil, fmt.Errorf("unsupported model number: %s", modelNumber)
	}

	responses, err := belkinExt.Scan(scanType, 5)
	fmt.Printf("%+v\n", responses)
	if err != nil {
		log.V("scan err: %s", err)
		return nil, err
	}

	devices := make([]gohome.Device, len(responses))
	for i, devInfo := range responses {
		err := devInfo.Load()

		if err != nil {
			// Keep going, try to get as many as we can
			log.V("failed to load device information: %s", err)
			continue
		}

		//fmt.Printf("%#v\n", response)
		//fmt.Printf("%#v\n", devInfo)

		cmdBuilder, ok := sys.Extensions.CmdBuilders[modelNumber]
		if !ok {
			return nil, fmt.Errorf("unsupported command builder ID: %s", modelNumber)
		}

		dev, _ := gohome.NewDevice(
			modelNumber,
			strings.Replace(devInfo.Scan.Location, "/setup.xml", "", -1),
			"",
			devInfo.FriendlyName,
			devInfo.ModelDescription,
			nil,
			false,
			cmdBuilder,
			nil,
			nil,
		)

		z := &zone.Zone{
			Address:     "",
			Name:        devInfo.FriendlyName,
			Description: devInfo.ModelDescription,
			DeviceID:    "",
			Type:        zone.ZTSwitch,
			Output:      zone.OTBinary,
		}
		dev.AddZone(z)
		devices[i] = *dev
	}

	return devices, nil
}