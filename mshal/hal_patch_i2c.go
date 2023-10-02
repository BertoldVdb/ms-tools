package mshal

func (h *HAL) patchI2CStart() error {
	addr := 0x3639
	if h.deviceType == 2107 {
		addr = 0x68bd
	} else if h.deviceType == 2109 {
		addr = 0x6a8c
	}

	_, err := h.PatchExecFunc(true, addr, PatchExecFuncRequest{})
	return err
}

func (h *HAL) patchI2CStop() error {
	addr := 0x3730
	if h.deviceType == 2107 {
		addr = 0x6b5b
	} else if h.deviceType == 2109 {
		addr = 0x6aba
	}
	_, err := h.PatchExecFunc(true, addr, PatchExecFuncRequest{})
	return err
}

func (h *HAL) patchI2CRead(ack bool) (uint8, error) {
	addr := 0x26cb
	if h.deviceType == 2109 || h.deviceType == 2107 {
		addr = h.patchCallAddrs[3]
	}
	r7 := byte(1)
	if ack {
		r7 = 0
	}
	resp, err := h.PatchExecFunc(true, addr, PatchExecFuncRequest{R7_A: r7})
	return resp.R7, err
}

func (h *HAL) patchI2CWrite(value uint8) (bool, error) {
	addr := 0x2126
	if h.deviceType == 2107 {
		addr = 0x5323
	} else if h.deviceType == 2109 {
		addr = 0x4648
	}
	resp, err := h.PatchExecFunc(true, addr, PatchExecFuncRequest{R7_A: value})
	if h.deviceType != 2106 {
		return resp.C, err
	}
	return resp.R7 > 0, err
}

func (h *HAL) I2CTransfer(addr uint8, wrBuf []byte, rdBuf []byte) (bool, error) {
	if !h.patchInstalled {
		return false, ErrorMissingFunction
	}

	if len(rdBuf) == 0 && len(wrBuf) == 0 {
		return true, nil

	}
	if len(wrBuf) > 0 {
		if err := h.patchI2CStart(); err != nil {
			return false, err
		}

		if ack, err := h.patchI2CWrite(addr << 1); !ack || err != nil {
			return false, err
		}

		for _, m := range wrBuf {
			if ack, err := h.patchI2CWrite(m); !ack || err != nil {
				return false, err
			}
		}
	}

	if len(rdBuf) > 0 {
		if err := h.patchI2CStart(); err != nil {
			return false, err
		}

		if ack, err := h.patchI2CWrite(addr<<1 | 1); !ack || err != nil {
			return false, err
		}

		for i := range rdBuf {
			value, err := h.patchI2CRead(i < len(rdBuf)-1)
			if err != nil {
				return false, err
			}
			rdBuf[i] = value
		}
	}

	if err := h.patchI2CStop(); err != nil {
		return false, err
	}

	return true, nil
}
