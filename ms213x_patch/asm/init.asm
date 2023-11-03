; Tell host about the existence of this patch

MOV  DPTR, #0x7B00
MOV  A, #'B'
MOVX @DPTR, A
INC  DPTR
MOV  A, #'V'
MOVX @DPTR, A
INC  DPTR
MOV  A, #'D'
MOVX @DPTR, A
INC  DPTR
MOV  A, #'B'
MOVX @DPTR, A
INC  DPTR
CLR  A
MOVX @DPTR, A
RET
