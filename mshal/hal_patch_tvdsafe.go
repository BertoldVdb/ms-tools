package mshal

type halPatchTVDMemoryRegion struct {
	hal *HAL
}

func (h halPatchTVDMemoryRegion) GetName() MemoryRegionNameType {
	return MemoryRegionRegisters2106TVD
}

func (h halPatchTVDMemoryRegion) GetLength() int {
	return 256
}

func (h halPatchTVDMemoryRegion) GetParent() (MemoryRegion, int) {
	return nil, 0
}

func (h halPatchTVDMemoryRegion) read(addr int, buf []byte) (int, error) {
	req := PatchExecFuncRequest{
		R7_A: uint8(addr),
	}

	resp, err := h.hal.PatchExecFunc(false, 0x3a33, req)
	if err != nil {
		return 0, err
	}

	buf[0] = resp.R7
	return 1, nil
}

func (h halPatchTVDMemoryRegion) write(addr int, buf []byte) (int, error) {
	req := PatchExecFuncRequest{
		R5:   uint8(buf[0]),
		R7_A: uint8(addr),
	}

	_, err := h.hal.PatchExecFunc(false, 0x3a17, req)
	if err != nil {
		return 0, err
	}

	return 1, nil
}

func (h halPatchTVDMemoryRegion) Access(write bool, addr int, buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	if write {
		return h.write(addr, buf)
	}

	return h.read(addr, buf)
}

func (h *HAL) patchMakeTVDRegion() MemoryRegion {
	if !h.patchInstalled {
		return nil
	}

	return regionWrapCompleteIO(halPatchTVDMemoryRegion{hal: h})
}
