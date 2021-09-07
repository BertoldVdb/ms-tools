// +build puregohid

package main

import (
	"errors"

	"github.com/BertoldVdb/ms-tools/gohid"
)

func OpenDevice() (gohid.HIDDevice, error) {
	if CLI.RawPath == "" {
		return nil, errors.New("RawPath must be specified when using pure GO HID")
	}

	return gohid.OpenHID(CLI.RawPath)
}

type ListHIDCmd struct {
}

func (l *ListHIDCmd) Run(c *Context) error {
	return errors.New("This command is not supported using pure GO HID")
}
