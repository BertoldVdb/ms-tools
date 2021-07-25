MOV A,  P3 ;Direction (1=input, 0=output)
ORL A,  R6
ANL A,  R7
MOV P3, A
MOV A,  P2 ;State (1=high, 0=low)
ORL A,  R4
ANL A,  R5
MOV P2, A
MOV R2, P2
MOV R3, P3
RET
