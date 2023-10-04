#!/bin/sh

(echo ".equ HID, 0x13"; cat hook.asm) > hook_2109.asm
(echo ".equ HID, 0x14"; cat hook.asm) > hook_2106.asm
as31 -Fbin hook_2109.asm
as31 -Fbin hook_2106.asm
rm hook_2109.asm
rm hook_2106.asm
as31 -Fbin gpio.asm 
as31 -Fbin code.asm
as31 -Fbin i2cRead2109.asm
as31 -Fbin i2cRead2107.asm
as31 -Fbin uart_tx.asm
