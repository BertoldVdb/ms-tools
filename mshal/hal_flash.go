package mshal

import (
	"encoding/binary"
)

func (h *HAL) ms2130enableSPI(enable bool) error {
	value := byte(0x00)
	if enable {
		if h.ms2130spiEnabled == 1 {
			return nil
		}

		/* Configure GPIO */
		output := byte(1<<2 | 1<<3 | 1<<4)
		input := byte(1 << 5)
		if _, _, err := h.GPIOUpdate(output, 0, output, input); err != nil {
			return err
		}

		value = byte(0x10)
	} else {
		if h.ms2130spiEnabled == 0 {
			return nil
		}
	}

	/* Configure pin mux */
	_, err := h.MemoryRegionGet(MemoryRegionRAM).Access(true, 0xf01f, []byte{value})

	if enable {
		h.ms2130spiEnabled = 1
	} else {
		h.ms2130spiEnabled = 0
	}

	return err
}

type romFlashMemoryRegion struct {
	hal *HAL

	flashReadBufferValid bool
	flashReadBufferPage  uint16
}

func (h *HAL) memoryRegionFlash() MemoryRegion {
	return regionWrapCompleteIO(&romFlashMemoryRegion{
		hal: h,
	})
}

func (r romFlashMemoryRegion) GetLength() int {
	return 0x10000
}

func (r *romFlashMemoryRegion) Access(write bool, addr int, buf []byte) (int, error) {
	if err := r.hal.ms2130enableSPI(true); err != nil {
		return 0, err
	}

	if write {
		r.flashReadBufferValid = false

		if addr == 0 {
			/* Erase the flash first */
			var out [8]byte
			out[0] = 0xfe
			if _, err := r.hal.ROMExchangeReport(out[:]); err != nil {
				return 0, err
			}
		}

		/* Setup flash write: f801aaaaaabbbb00 (aaaaaa=address, bbbb=blocksize) *
		 * Write to the buffer: f800cccccccccc (cccccccccccc=data) */

		if len(buf) > 256 {
			buf = buf[:256]
		}

		var out [8]byte
		out[0] = 0xf8
		out[1] = 0x01
		out[2] = byte(addr >> 16)
		out[3] = byte(addr >> 8)
		out[4] = byte(addr >> 0)
		binary.BigEndian.PutUint16(out[5:], uint16(len(buf)))

		if _, err := r.hal.ROMExchangeReport(out[:]); err != nil {
			return 0, err
		}

		written := 0
		for len(buf) > 0 {
			out[0] = 0xf8
			out[1] = 0x00

			n := copy(out[2:], buf)

			if _, err := r.hal.ROMExchangeReport(out[:]); err != nil {
				return written, err
			}

			written += n
			buf = buf[n:]
		}

		return written, nil
	}

	flashPage := uint16(addr >> 8)
	flashOffset := uint8(addr & 0xff)

	/* Read from flash to buffer: f701aaaaaabbbb00 (aaaaaa=addr, bbbb=len to read)
	 * Read from buffer to hostt: f700000000aaaa00 (aaaa=offset) */

	if !r.flashReadBufferValid || r.flashReadBufferPage != flashPage {
		var out [8]byte
		out[0] = 0xf7
		out[1] = 0x01
		binary.BigEndian.PutUint16(out[2:], flashPage)
		binary.BigEndian.PutUint16(out[5:], 256)

		if _, err := r.hal.ROMExchangeReport(out[:]); err != nil {
			return 0, err
		}

		r.flashReadBufferValid = true
		r.flashReadBufferPage = flashPage
	}

	var out [8]byte
	out[0] = 0xf7
	binary.BigEndian.PutUint16(out[5:], uint16(flashOffset))
	in, err := r.hal.ROMExchangeReport(out[:])
	if err != nil {
		return 0, err
	}

	maxLen := 0x100 - int(flashOffset)
	if len(buf) > maxLen {
		buf = buf[:maxLen]
	}

	return copy(buf, in), nil
}

func (r romFlashMemoryRegion) GetParent() (MemoryRegion, int) {
	return nil, 0
}
func (r romFlashMemoryRegion) GetName() MemoryRegionNameType {
	return MemoryRegionFLASH
}

func (r romFlashMemoryRegion) GetAlignment() int {
	return 1
}
