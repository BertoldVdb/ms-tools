.FLAG PIN, P2.4
.EQU  BAUD0, 126
.EQU  BAUD1, 127

    PUSH BAUD0
    PUSH BAUD1

    MOVX A, @DPTR
    INC  DPTR
    MOV  BAUD0, A
    MOVX A, @DPTR
    MOV  BAUD1, A

more:
    MOV  R0, #8    ;Send 8 bits from A
    INC  DPTR

    MOV  A, R6
    RRC  A
    MOV  PIN,C     ;Send start bit

    MOV  R1, BAUD0 ;Delay between bits
    MOV  R2, BAUD1
d1:
    DJNZ R1, d1
    DJNZ R2, d1

    MOVX A, @DPTR

bits:
    RRC  A         ;Rotate bits via carry
    MOV  PIN,C     ;Output bits

    MOV  R1, BAUD0 ;Delay between bits
    MOV  R2, BAUD1
d2:
    DJNZ R1, d2
    DJNZ R2, d2
    
    DJNZ R0, bits  ;Send all bits

    MOV   A, R6
    RLC   A
    MOV   PIN,C     ;Send stop bit
    
    MOV  R1, BAUD0 ;Delay between bits
    MOV  R2, BAUD1
d3:
    DJNZ R1, d3
    DJNZ R2, d3

    DJNZ R7, more  ;Send all characters

    POP  BAUD0
    POP  BAUD1
    RET
