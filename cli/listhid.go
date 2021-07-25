package main

import (
	"fmt"

	"github.com/sstallion/go-hid"
)

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
}
