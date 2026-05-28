package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// RestartGemini attempts to kill the parent process and start a new instance of gemini
func RestartGemini(pid int, delay time.Duration) {
	go func() {
		fmt.Fprintf(os.Stderr, "[Restart Helper] Waiting %v before restart...\n", delay)
		time.Sleep(delay)

		// 1. Kill the old process
		fmt.Fprintf(os.Stderr, "[Restart Helper] Terminating PID: %d...\n", pid)
		process, err := os.FindProcess(pid)
		if err == nil {
			_ = process.Signal(os.Interrupt) // Try SIGINT first
			time.Sleep(500 * time.Millisecond)
			_ = process.Kill()
		}

		// 2. Start new instance
		fmt.Fprintf(os.Stderr, "[Restart Helper] Starting new Gemini CLI...\n")
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("cmd.exe", "/c", "start", "gemini")
		case "darwin":
			cmd = exec.Command("open", "-a", "Terminal", "gemini") // Or just "gemini" if it's a CLI
		default:
			cmd = exec.Command("gemini")
		}

		if cmd != nil {
			err := cmd.Start()
			if err != nil {
				fmt.Fprintf(os.Stderr, "[Restart Helper] Failed to start new instance: %v\n", err)
			}
		}
	}()
}
