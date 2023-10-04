.FLAG PIN, P2.2
.FLAG DIR, P3.2
.EQU  BAUD0, 120
.EQU  BAUD1, 121
.EQU  PDIR,  122

    MOV BAUD0, R5 ;Save baud rate param
    MOV BAUD1, R6
    MOV PDIR,  P3
    CLR DIR

more:
    MOVX A, @DPTR
    INC  DPTR

    MOV  R0, #8    ;Send 8 bits from A
    CLR  PIN       ;Send start bit

    MOV  R5, BAUD0 ;Delay between bits
    MOV  R6, BAUD1
d1:
    DJNZ R5, d1
    DJNZ R6, d1 

bits:
    RRC  A         ;Rotate bits via carry
    MOV  PIN,C     ;Output bits

    MOV  R5, BAUD0 ;Delay between bits
    MOV  R6, BAUD1
d2:
    DJNZ R5, d2
    DJNZ R6, d2 
    
    DJNZ R0, bits  ;Send all bits

    RRC  A         ;Not needed, but makes timing equal
    SETB PIN       ;Send stop bit
    
    MOV  R5, BAUD0 ;Delay between bits
    MOV  R6, BAUD1
d3:
    DJNZ R5, d3
    DJNZ R6, d3 

    DJNZ R7, more  ;Send all characters

    MOV  P3, PDIR

    RET
