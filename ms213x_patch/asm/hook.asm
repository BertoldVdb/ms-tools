hookRun:
	MOV   DPTR, #0x12b3

	MOVX  A, @DPTR
	XRL   A, R0
	JNZ   hookRet
	
	LCALL hookWork

    MOV   DPTR, #0x12b3+1
	MOVX  @DPTR, A
	INC   DPTR
	MOV   A, R2
	MOVX  @DPTR, A
	INC   DPTR
	MOV   A, R3
	MOVX  @DPTR, A
	INC   DPTR
	MOV   A, R4
	MOVX  @DPTR, A
	INC   DPTR
	MOV   A, R5
	MOVX  @DPTR, A
	INC   DPTR
	MOV   A, R6
	MOVX  @DPTR, A
	INC   DPTR
	MOV   A, R7
	MOVX  @DPTR, A
	MOV   DPTR, #0x12b3
	MOV   A,     #0xFF
	RLC   A
	MOVX  @DPTR, A

	SETB EA

hookRet:
	RET

hookWork:
    CLR EA

	MOV  DPTR, #0x12b3+1
	MOVX A, @DPTR
	MOV  1, A

	INC  DPTR
	MOVX A, @DPTR
	MOV  0, A

	PUSH 0
	PUSH 1

	INC DPTR
	MOVX A, @DPTR
	MOV  R3,  A

	INC  DPTR
	MOVX A, @DPTR
	MOV  R4, A

	INC  DPTR
	MOVX A, @DPTR
	MOV  R5, A

	INC  DPTR
	MOVX A, @DPTR
	MOV  R6, A

	INC  DPTR
	MOVX A, @DPTR
	MOV  R7,  A

    RRC  A
	MOV  A,  R7

	MOV DPH, R3
	MOV DPL, R4

	RET ;Call address in HID+1


