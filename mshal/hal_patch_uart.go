package mshal

import "encoding/binary"

func (h *HAL) UARTTransmit(baud int, data []byte, invert bool) error {
	if _, _, err := h.GPIOUpdate(0, 0, 1>>4, 0); err != nil {
		return err
	}

	region := h.MemoryRegionGet(MemoryRegionUserRAM)
	parent, addr := region.GetParent()

	params := make([]byte, 2+len(data))
	addr += region.GetLength() - len(params)

	binary.LittleEndian.PutUint16(params[:], uint16((1.0/float64(baud))/108.125e-9))
	params[0] += 1
	params[1] += 1

	copy(params[2:], data)
	ssbit := byte(0x80)
	if invert {
		ssbit = 0x01
		for i := range data {
			params[2+i] ^= 0xFF
		}
	}

	if _, err := parent.Access(true, addr, params); err != nil {
		return err
	}

	_, err := h.PatchExecFunc(false, h.patchCallAddrs[4], PatchExecFuncRequest{DPTR: uint16(addr), R6: ssbit, R7_A: uint8(len(data))})
	return err
}
