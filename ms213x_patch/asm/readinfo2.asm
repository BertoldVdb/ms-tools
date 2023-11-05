; Resolution info
MOV  DPTR, #0xe184
MOVX A, @DPTR
MOV  R2, A

INC  DPTR
MOVX A, @DPTR
MOV  R3, A

MOV  DPTR, #0xe18c
MOVX A, @DPTR
MOV  R4, A

INC  DPTR
MOVX A, @DPTR
MOV  R5, A

; Signal info
MOV  DPTR, #0xf6e9
MOVX A, @DPTR
MOV  R6, A

; Frame counter
MOV  DPTR, #0x7b16
MOVX A, @DPTR

RET
