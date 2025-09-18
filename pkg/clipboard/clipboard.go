package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Copy copies text to the system clipboard
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

// copyMacOS copies text to clipboard on macOS using pbcopy
func copyMacOS(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// copyLinux copies text to clipboard on Linux using xclip or xsel
func copyLinux(text string) error {
	// Try xclip first
	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Try xsel as fallback
	if _, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command("xsel", "--clipboard", "--input")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}

	return fmt.Errorf("no clipboard utility found (install xclip or xsel)")
}

// copyWindows copies text to clipboard on Windows using clip
func copyWindows(text string) error {
	cmd := exec.Command("clip")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// Paste reads text from the system clipboard
func Paste() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return pasteMacOS()
	case "linux":
		return pasteLinux()
	case "windows":
		return pasteWindows()
	default:
		return "", fmt.Errorf("clipboard operations not supported on %s", runtime.GOOS)
	}
}

// pasteMacOS reads text from clipboard on macOS using pbpaste
func pasteMacOS() (string, error) {
	cmd := exec.Command("pbpaste")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(output), "\n"), nil
}

// pasteLinux reads text from clipboard on Linux using xclip or xsel
func pasteLinux() (string, error) {
	// Try xclip first
	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard", "-out")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimRight(string(output), "\n"), nil
		}
	}

	// Try xsel as fallback
	if _, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command("xsel", "--clipboard", "--output")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimRight(string(output), "\n"), nil
		}
	}

	return "", fmt.Errorf("no clipboard utility found (install xclip or xsel)")
}

// pasteWindows reads text from clipboard on Windows
func pasteWindows() (string, error) {
	// Windows doesn't have a simple command-line paste utility
	// This would require Windows API calls or PowerShell
	return "", fmt.Errorf("paste not implemented for Windows")
}
