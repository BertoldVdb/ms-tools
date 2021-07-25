package mshal

func (h *HAL) GPIOUpdate(stateSet byte, stateClear byte, outputSet byte, outputClear byte) (byte, byte, error) {
	if !h.patchInstalled {
		return 0, 0, ErrorMissingFunction
	}

	var req patchExecFuncRequest
	req.R4 = stateSet
	req.R5 = ^stateClear
	req.R6 = outputClear
	req.R7_A = ^outputSet
	resp, err := h.patchExecFunc(true, h.patchCallAddrs[1], req)

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
