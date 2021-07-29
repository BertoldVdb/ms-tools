package mshal

func (h *HAL) patchReadCode(addr int) (byte, error) {
	resp, err := h.PatchExecFunc(true, h.patchCallAddrs[2], PatchExecFuncRequest{DPTR: uint16(addr)})
	if err != nil {
		return 0, err
	}
	return resp.A, nil
}

type halPatchCodeMemoryRegion struct {
	hal *HAL
}

func (h halPatchCodeMemoryRegion) GetName() MemoryRegionNameType {
	return MemoryRegionCODE
}

func (h halPatchCodeMemoryRegion) GetLength() int {
	return 0x10000
}

func (h halPatchCodeMemoryRegion) GetParent() (MemoryRegion, int) {
	return nil, 0
}

func (h halPatchCodeMemoryRegion) Access(write bool, addr int, buf []byte) (int, error) {
	if write {
		return 0, ErrorWriteNotAllowed
	}

	if len(buf) == 0 {
		return 0, nil
	}

	value, err := h.hal.patchReadCode(addr)
	if err != nil {
		return 0, err
	}

	buf[0] = value
	return 1, nil
}

func (h *HAL) patchMakeCodeRegion() MemoryRegion {
	if !h.patchInstalled {
		return nil
	}

	return regionWrapCompleteIO(halPatchCodeMemoryRegion{hal: h})
}
