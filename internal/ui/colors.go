package ui

import (
	"fmt"
	"os"
	"runtime"
)

// Color codes for terminal output.
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
)

var colorEnabled = true

func init() {
	// Disable colors on Windows (unless WT_SESSION is set indicating Windows Terminal)
	if runtime.GOOS == "windows" {
		if os.Getenv("WT_SESSION") == "" {
			colorEnabled = false
		}
	}
	if os.Getenv("NO_COLOR") != "" {
		colorEnabled = false
	}
}

// SetColorEnabled allows toggling color output.
func SetColorEnabled(enabled bool) {
	colorEnabled = enabled
}

// Colorize wraps text with the given color code.
func Colorize(color, text string) string {
	if !colorEnabled {
		return text
	}
	return color + text + Reset
}

// Success prints a green checkmark message.
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(Colorize(Green, "✓ ") + msg)
}

// Warn prints a yellow warning message.
func Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(Colorize(Yellow, "⚠ ") + msg)
}

// Error prints a red error message.
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, Colorize(Red, "✗ ")+msg)
}

// Info prints a blue info message.
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(Colorize(Blue, "ℹ ") + msg)
}

// Header prints a bold header.
func Header(text string) {
	fmt.Println(Colorize(Bold, text))
}
