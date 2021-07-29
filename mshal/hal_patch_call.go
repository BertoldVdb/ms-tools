package mshal

import (
	"encoding/hex"
	"errors"
	"time"
)

func (h *HAL) patchExchangeReport(out [9]byte) ([9]byte, error) {
	var in [9]byte

	if h.config.LogFunc != nil {
		h.config.LogFunc(3, "PatchOut: %s", hex.EncodeToString(out[:]))
	}

	if _, err := h.dev.SendFeatureReport(out[:]); err != nil {
		return in, err
	}

	timeout := time.Now().Add(3 * time.Second)

	for time.Now().Before(timeout) {
		_, err := h.dev.GetFeatureReport(in[:])
		if err != nil {
			return in, err
		}

		if in[1]&0xFE == 0xFE {
			if h.config.LogFunc != nil {
				h.config.LogFunc(3, "PatchIn:  %s", hex.EncodeToString(in[:]))
			}

			return in, nil
		}
	}

	return in, ErrorTimeout
}

type PatchExecFuncResponse struct {
	A  byte
	R2 byte
	R3 byte
	R4 byte
	R5 byte
	R6 byte
	R7 byte
	C  bool
}

type PatchExecFuncRequest struct {
	DPTR uint16
	R3   byte
	R4   byte
	R5   byte
	R6   byte
	R7_A byte
}

func (h *HAL) PatchExecFunc(inIRQ bool, addr int, req PatchExecFuncRequest) (PatchExecFuncResponse, error) {
	var response PatchExecFuncResponse

	if !h.patchInstalled {
		return response, ErrorMissingFunction
	}

	if req.DPTR != 0 && (req.R4 != 0 || req.R3 != 0) {
		return response, errors.New("Can't set both DPTR and R3/R4")
	}

	var out [9]byte
	out[1] = 0xef
	if inIRQ {
		out[1] = 0xee
	}
	out[2] = byte(addr >> 8)
	out[3] = byte(addr)
	if req.DPTR != 0 {
		out[4] = byte(req.DPTR >> 8)
		out[5] = byte(req.DPTR)
	} else {
		out[4] = req.R3
		out[5] = req.R4
	}
	out[6] = req.R5
	out[7] = req.R6
	out[8] = req.R7_A

	in, err := h.patchExchangeReport(out)
	if err != nil {
		return response, err
	}

	response.A = in[2]
	response.R2 = in[3]
	response.R3 = in[4]
	response.R4 = in[5]
	response.R5 = in[6]
	response.R6 = in[7]
	response.R7 = in[8]
	response.C = in[1]&1 > 0
	return response, nil
}

func (h *HAL) PatchCodeBlobGetAddress(index int) int {
	if index < 0 {
		return 0
	}

	index += h.patchCallAddrsExternalStart
	if index >= len(h.patchCallAddrs) {
		return 0
	}

	return h.patchCallAddrs[index]
}
