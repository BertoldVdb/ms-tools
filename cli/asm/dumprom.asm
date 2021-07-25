.equ CommAddr, 0xDEAD

; Wait for number of bytes to be written
MOV DPTR, #CommAddr
MOVX A, @DPTR
JNZ work
RET
work:
MOV R7, A

; Read source address
INC DPTR
MOVX A, @DPTR
MOV R0, A
INC DPTR
MOVX A, @DPTR
MOV R1, A

; Read dest address
INC DPTR
MOVX A, @DPTR
MOV R2, A
INC DPTR
MOVX A, @DPTR
MOV R3, A

; Clear code index
MOV R4, #0

dump:
; Read from code
MOV DPH, R0
MOV DPL, R1

MOV A, R4
MOVC A, @A+DPTR
; Write to XDATA
MOV DPH, R2
MOV DPL, R3
MOVX @DPTR, A

; Update indices
INC R3
INC R4
DJNZ R7, dump

; Signal completion
MOV DPTR, #CommAddr
CLR A
MOVX @DPTR, A
RET

