package mshal

func (h *HAL) GPIOUpdate(stateSet byte, stateClear byte, outputSet byte, outputClear byte) (byte, byte, error) {
	if !h.patchInstalled {
		if sfr := h.MemoryRegionGet(MemoryRegionSFR); sfr != nil {
			return h.gpioUpdateSFR(sfr, stateSet, stateClear, outputSet, outputClear)
		}
		return 0, 0, ErrorMissingFunction
	}

	var req PatchExecFuncRequest
	req.R4 = stateSet
	req.R5 = ^stateClear
	req.R6 = outputClear
	req.R7_A = ^outputSet
	resp, err := h.PatchExecFunc(true, h.patchCallAddrs[1], req)

	return resp.R2, ^resp.R3, err
}

func (h *HAL) GPIOWrite(index int, value bool) error {
	if value {
		return h.GPIOSet(index)
	}
	return h.GPIOClear(index)
}

func (h *HAL) GPIOSet(index int) error {
	_, _, err := h.GPIOUpdate(1<<index, 0, 1<<index, 0)
	return err
}

func (h *HAL) GPIOClear(index int) error {
	_, _, err := h.GPIOUpdate(0, 1<<index, 1<<index, 0)
	return err
}

func (h *HAL) GPIORead(index int) (bool, error) {
	p2, _, err := h.GPIOUpdate(0, 0, 0, 1<<index)
	return p2&(1<<index) > 0, err
}

func (h *HAL) gpioUpdateSFR(sfr MemoryRegion, stateSet byte, stateClear byte, outputSet byte, outputClear byte) (byte, byte, error) {
	if err := h.ms2130enableSPI(false); err != nil {
		return 0, 0, err
	}

	/* P3 = 0xB0, P2 = 0xA0 */
	var P3, P2 [1]byte
	if _, err := sfr.Access(false, 0xB0-0x80, P3[:]); err != nil {
		return 0, 0, err
	}
	if _, err := sfr.Access(false, 0xA0-0x80, P2[:]); err != nil {
		return 0, 0, err
	}

	P3[0] |= outputClear
	P3[0] &= ^outputSet

	P2[0] |= stateSet
	P2[0] &= ^stateClear

	if _, err := sfr.Access(true, 0xB0-0x80, P3[:]); err != nil {
		return 0, 0, err
	}
	if _, err := sfr.Access(true, 0xA0-0x80, P2[:]); err != nil {
		return 0, 0, err
	}

	return P2[0], P3[0], nil
}
