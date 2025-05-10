package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenBrowser abre la URL en el navegador por defecto.
func OpenBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		fmt.Println("Por favor, abre esta URL manualmente:", url)
		return
	}
	exec.Command(cmd, args...).Start()
}
