package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/BertoldVdb/ms-tools/mshal"
	"github.com/inancgumus/screen"
)

type MEMIOListRegions struct {
}

func (l *MEMIOListRegions) Run(c *Context) error {
	var regions []mshal.MemoryRegion
	for _, m := range c.hal.MemoryRegionList() {
		regions = append(regions, c.hal.MemoryRegionGet(m))
	}

	fmt.Printf("Region       |     Length | Parent (%s)\n", c.hal.GetDeviceType())

	for _, m := range regions {
		parent, offset := mshal.RecursiveGetParentAddress(m, 0)
		fmt.Printf("%-13s|      %5d |", m.GetName(), m.GetLength())
		if parent != m {
			fmt.Printf(" %s.%04X", parent.GetName(), offset)
		}
		fmt.Printf("\n")
	}
	return nil
}

type Region struct {
	Region string `arg name:"region" help:"Memory region to access."`
	Addr   int    `arg name:"addr" help:"Addresses to access." type:"int"`
}

type MEMIOReadCmd struct {
	Loop     int    `optional help:"0=Perform once, 1=Mark changes since start, 2=Mark changes since previous iteration."`
	Filename string `optional help:"File to write dump to."`

	Region Region `embed`
	Amount int    `arg name:"amount" help:"Number of bytes to read, omit for maximum." optional default:"0"`
}

func (l *MEMIOReadCmd) Run(c *Context) error {
	if l.Loop < 0 || l.Loop > 2 {
		return errors.New("Loop flag out of range")
	}

	region := c.hal.MemoryRegionGet(mshal.MemoryRegionNameType(l.Region.Region))
	if region == nil {
		return errors.New("Invalid memory region")
	}

	if l.Amount == 0 {
		l.Amount = region.GetLength()
	}

	var oldBuf []byte
	var mark []bool
	for {
		startTime := time.Now()
		if l.Loop == 2 || mark == nil {
			mark = make([]bool, l.Amount)
		}

		buf := make([]byte, l.Amount)
		n, err := region.Access(false, l.Region.Addr, buf)
		if err != nil {
			return fmt.Errorf("Read error: %s", err.Error())
		}
		buf = buf[:n]

		if l.Filename != "" {
			err := ioutil.WriteFile(l.Filename, buf, 0644)
			return err
		}

		if l.Amount == 1 {
			if len(buf) < 1 {
				return errors.New("0 bytes returned")
			}
			fmt.Printf("0x%02x\n", buf[0])
		} else {
			if l.Loop != 0 {
				screen.Clear()
				screen.MoveTopLeft()
				if oldBuf != nil {
					for i, m := range oldBuf {
						if m != buf[i] {
							mark[i] = true
						}
					}
				}
			}
			fmt.Println(hexdump(l.Region.Addr, buf, mark))
		}

		oldBuf = buf

		if l.Loop == 0 {
			break
		}
		d := time.Now().Sub(startTime)
		td := 200 * time.Millisecond
		if d < td {
			time.Sleep(td - d)
		}
	}

	return nil
}

type MEMIOWriteCmd struct {
	Zone  Region `embed`
	Value int    `arg name:"value" help:"Value to write." type:"int"`
}

func (w MEMIOWriteCmd) Run(c *Context) error {
	region := c.hal.MemoryRegionGet(mshal.MemoryRegionNameType(w.Zone.Region))
	if region == nil {
		return errors.New("Invalid memory region")
	}

	var value [1]byte
	value[0] = byte(w.Value)
	_, err := region.Access(true, w.Zone.Addr, value[:])
	return err
}

type MEMIOWriteFileCmd struct {
	Region   Region `embed`
	Filename string `arg name:"filename" help:"File to read data from."`

	Verify bool `optional name:"verify" help:"Read and verify written file."`
}

func (w MEMIOWriteFileCmd) Run(c *Context) error {
	data, err := ioutil.ReadFile(w.Filename)
	if err != nil {
		return err
	}

	region := c.hal.MemoryRegionGet(mshal.MemoryRegionNameType(w.Region.Region))
	if region == nil {
		return errors.New("Invalid memory region")
	}

	n, err := region.Access(true, w.Region.Addr, data)
	if n > 0 {
		fmt.Printf("Wrote %d bytes to %s:%04x.\n", n, w.Region.Region, w.Region.Addr)
	}

	if w.Verify {
		readback := make([]byte, len(data))
		_, err := region.Access(false, w.Region.Addr, readback)
		if err != nil {
			return err
		}

		if !bytes.Equal(readback, data) {
			return errors.New("Failed to verify write")
		}

		fmt.Println("Verification OK.")
	}

	return err
}
