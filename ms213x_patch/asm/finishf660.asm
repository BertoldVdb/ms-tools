INC  DPTR
MOVX @DPTR, A

MOV  DPTR, #0x7b10
MOVX A, @DPTR
MOV  R7, A

INC  DPTR
MOVX A, @DPTR
MOV  R6, A

INC  DPTR
MOVX A, @DPTR
MOV  R5, A

INC  DPTR
MOVX A, @DPTR
MOV  R4, A

MOV  DPTR, #0xf660
MOV  A, R7
MOVX @DPTR, A

INC  DPTR
MOV  A, R6
MOVX @DPTR, A

INC  DPTR
MOV  A, R5
MOVX @DPTR, A

INC  DPTR
MOV  A, R4
MOVX @DPTR, A

RET
