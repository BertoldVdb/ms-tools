// +build !linux

package gohid

import "errors"

func openHIDInternal(path string) (HIDDevice, error) {
	return nil, errors.New("Platform is not supported")
}
