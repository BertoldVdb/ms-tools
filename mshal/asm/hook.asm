hookRun:
	MOV   A, HID
	XRL   A, R0
	JNZ   hookRet
	
	LCALL hookWork

	MOV   HID+1, A
	MOV   HID+2, R2
	MOV   HID+3, R3
	MOV   HID+4, R4
	MOV   HID+5, R5
	MOV   HID+6, R6
	MOV   HID+7, R7
	MOV   A,     #0xFF
	RLC   A
	MOV   HID,   A

hookRet:
	RET

hookWork:
	MOV  DPH, HID+3
	MOV  DPL, HID+4
	MOV  R3,  HID+3
	MOV  R4,  HID+4
	MOV  R5,  HID+5
	MOV  R6,  HID+6
	MOV  R7,  HID+7
	MOV  A,   R7
        RRC  A
	MOV  A,   R7

	PUSH      HID+2
	PUSH      HID+1	
	RET ;Call address in HID+1


