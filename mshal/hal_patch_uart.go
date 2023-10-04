package mshal

import (
	"encoding/binary"
)

func (h *HAL) UARTTransmit(baud int, buf []byte) error {
	region := h.MemoryRegionGet(MemoryRegionUserRAM)
	parent, addr := region.GetParent()
	addr += region.GetLength() - len(buf)

	if _, err := parent.Access(true, addr, buf); err != nil {
		return err
	}

	var div [2]byte
	binary.LittleEndian.PutUint16(div[:], uint16((1.0/float64(baud))/108.125e-9))

	_, err := h.PatchExecFunc(false, h.patchCallAddrs[4], PatchExecFuncRequest{DPTR: uint16(addr), R5: div[0] + 1, R6: div[1] + 1, R7_A: uint8(len(buf))})
	return err
}
