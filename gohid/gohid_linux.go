// +build linux

package gohid

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type HIDRaw struct {
	dev *os.File
}

func openHIDInternal(path string) (HIDDevice, error) {
	dev, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return &HIDRaw{
		dev: dev,
	}, nil
}

var ErrorTooLong = errors.New("Transfer is too long")

/*
 HIDIOCSFEATURE(0) = C0004806
 HIDIOCSFEATURE(9) = C0094806
 HIDIOCGFEATURE(0) = C0004807
 HIDIOCGFEATURE(9) = C0094807
*/

func (h *HIDRaw) SendFeatureReport(b []byte) (int, error) {
	var tmp [1024]byte

	if len(b) > len(tmp) {
		return 0, ErrorTooLong
	}

	copy(tmp[:], b)

	_, _, errno := unix.Syscall(
		syscall.SYS_IOCTL,
		uintptr(h.dev.Fd()),
		uintptr(uint32(0xC0004806)|uint32(len(b)<<16)),
		uintptr(unsafe.Pointer(&tmp)),
	)

	runtime.KeepAlive(tmp)

	if errno != 0 {
		return 0, os.NewSyscallError("SendFeatureReport", fmt.Errorf("%d", int(errno)))
	}

	return len(b), nil
}

func (h *HIDRaw) GetFeatureReport(b []byte) (int, error) {
	var tmp [256]byte

	if len(b) > len(tmp) {
		return 0, ErrorTooLong
	}

	_, _, errno := unix.Syscall(
		syscall.SYS_IOCTL,
		uintptr(h.dev.Fd()),
		uintptr(uint32(0xC0004807)|uint32(len(b)<<16)),
		uintptr(unsafe.Pointer(&tmp)),
	)

	if errno != 0 {
		return 0, os.NewSyscallError("GetFeatureReport", fmt.Errorf("%d", int(errno)))
	}

	copy(b, tmp[:])

	return len(b), nil
}

func (h *HIDRaw) Close() error {
	return h.dev.Close()
}
