// +build !puregohid

package main

import (
	"fmt"
	"os"

	"errors"

	"github.com/BertoldVdb/ms-tools/gohid"
	"github.com/sstallion/go-hid"
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

func OpenDevice() (gohid.HIDDevice, error) {
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

type ListHIDCmd struct {
}

func (l *ListHIDCmd) Run(c *Context) error {
	return SearchDevice(func(info *hid.DeviceInfo) error {
		fmt.Printf("%s: ID %04x:%04x %s %s\n",
			info.Path, info.VendorID, info.ProductID, info.MfrStr, info.ProductStr)
		fmt.Println("Device Information:")
		fmt.Printf("\tPath         %s\n", info.Path)
		fmt.Printf("\tVendorID     %04x\n", info.VendorID)
		fmt.Printf("\tProductID    %04x\n", info.ProductID)
		fmt.Printf("\tSerialNbr    %s\n", info.SerialNbr)
		fmt.Printf("\tReleaseNbr   %x.%x\n", info.ReleaseNbr>>8, info.ReleaseNbr&0xff)
		fmt.Printf("\tMfrStr       %s\n", info.MfrStr)
		fmt.Printf("\tProductStr   %s\n", info.ProductStr)
		fmt.Printf("\tUsagePage    %#x\n", info.UsagePage)
		fmt.Printf("\tUsage        %#x\n", info.Usage)
		fmt.Printf("\tInterfaceNbr %d\n", info.InterfaceNbr)
		fmt.Println()

		return nil
	})
	return nil
}
