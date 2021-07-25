package main

import (
	"os"

	"github.com/sstallion/go-hid"
    "errors"
)

func SearchDevice(foundHandler func(info *hid.DeviceInfo) error) error {
	return hid.Enumerate(uint16(CLI.VID), uint16(CLI.PID), func(info *hid.DeviceInfo) error {
		if CLI.Serial != "" && info.SerialNbr != CLI.Serial {
			return nil
		}
		if CLI.RawPath != "" && info.Path != CLI.RawPath {
			return nil
		}

		return foundHandler(info)
	})
}

func OpenDevice() (*hid.Device, error) {
	var dev *hid.Device
	err := SearchDevice(func(info *hid.DeviceInfo) error {
		d, err := hid.Open(info.VendorID, info.ProductID, info.SerialNbr)
		if err == nil {
			dev = d
			return errors.New("Done") 
		}
		return err
	})
	if dev != nil {
		return dev, nil
	}
	if err == nil {
		err = os.ErrNotExist
	}

	return nil, err
}
