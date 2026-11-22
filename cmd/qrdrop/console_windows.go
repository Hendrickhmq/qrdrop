//go:build windows

// Este arquivo só é compilado no Windows (a linha //go:build lá em cima). Nos
// outros sistemas vale console_other.go, com as mesmas funções.

package main

import (
	"os"
	"syscall"
)

// Funções da API do Windows (kernel32.dll) que o pacote syscall não traz prontas.
var (
	kernel32       = syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode = kernel32.NewProc("SetConsoleMode")
)

// enableVirtualTerminalProcessing é o ENABLE_VIRTUAL_TERMINAL_PROCESSING da
// documentação da Microsoft: com ele o console entende cores ANSI.
const enableVirtualTerminalProcessing = 0x0004

// enableANSI liga as cores ANSI no console (Windows 10 em diante). Devolve
// false se a saída não é um console (foi redirecionada para um arquivo) ou se
// o Windows recusou; aí o QR sai sem cores.
func enableANSI() bool {
	h := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	if syscall.GetConsoleMode(h, &mode) != nil {
		return false
	}
	ok, _, _ := setConsoleMode.Call(uintptr(h), uintptr(mode|enableVirtualTerminalProcessing))
	return ok != 0
}
