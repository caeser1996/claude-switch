package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SelectOption represents a choice in the interactive selector.
type SelectOption struct {
	Label       string
	Description string
	Value       string
}

// Select shows an interactive numbered list and lets the user pick.
// Returns the selected option's Value.
func Select(prompt string, options []SelectOption) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options to select from")
	}

	fmt.Println()
	Header(prompt)
	fmt.Println()

	for i, opt := range options {
		num := fmt.Sprintf("  %d)", i+1)
		label := opt.Label
		if opt.Description != "" {
			label += Colorize(Gray, "  "+opt.Description)
		}
		fmt.Printf("%s %s\n", Colorize(Cyan, num), label)
	}

	fmt.Println()
	fmt.Print(Colorize(Cyan, "  Enter number (1-"+fmt.Sprintf("%d", len(options))+"): "))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("cannot read input: %w", err)
	}

	input = strings.TrimSpace(input)
	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err != nil || idx < 1 || idx > len(options) {
		return "", fmt.Errorf("invalid selection: %q", input)
	}

	return options[idx-1].Value, nil
}

// Confirm asks a yes/no question. Returns true for yes.
func Confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// ReadPassword reads a line from stdin (no echo masking in this simple version).
func ReadPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}
