package shared

import (
	"os"
	"os/exec"
	"runtime"
)

type ContentType int

const (
	MESSAGE ContentType = iota
	INPUT
)

type Message struct {
	Content string
	Type    ContentType
}

func ClearTerminal() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	case "darwin", "linux":
		cmd = exec.Command("clear")
	default:
		// Unsupported operating system
		return
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}
