package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Copy copies text to the system clipboard.
func Copy(text string) error {
	switch runtime.GOOS {
	case "darwin":
		return copyMacOS(text)
	case "linux":
		return copyLinux(text)
	case "windows":
		return copyWindows(text)
	default:
		return fmt.Errorf("clipboard operations not supported on %s", runtime.GOOS)
	}
}

// copyMacOS copies text to clipboard on macOS using pbcopy.
func copyMacOS(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// copyLinux copies text to clipboard on Linux using xclip or xsel.
func copyLinux(text string) error {
	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	if _, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command("xsel", "--clipboard", "--input")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}

	return fmt.Errorf("no clipboard utility found (install xclip or xsel)")
}

// copyWindows copies text to clipboard on Windows using clip.
func copyWindows(text string) error {
	cmd := exec.Command("clip")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}