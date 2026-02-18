package ui

import (
	"fmt"
	"strings"
)

// Table renders a simple formatted table.
type Table struct {
	Headers []string
	Rows    [][]string
}

// NewTable creates a new table with the given headers.
func NewTable(headers ...string) *Table {
	return &Table{
		Headers: headers,
	}
}

// AddRow adds a row to the table.
func (t *Table) AddRow(cols ...string) {
	t.Rows = append(t.Rows, cols)
}

// Render prints the table to stdout.
func (t *Table) Render() {
	if len(t.Headers) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(t.Headers))
	for i, h := range t.Headers {
		widths[i] = len(h)
	}
	for _, row := range t.Rows {
		for i, col := range row {
			if i < len(widths) && len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	// Print header
	headerLine := ""
	separatorLine := ""
	for i, h := range t.Headers {
		if i > 0 {
			headerLine += "  "
			separatorLine += "  "
		}
		headerLine += Colorize(Bold, padRight(h, widths[i]))
		separatorLine += strings.Repeat("─", widths[i])
	}
	fmt.Println(headerLine)
	fmt.Println(Colorize(Gray, separatorLine))

	// Print rows
	for _, row := range t.Rows {
		line := ""
		for i, col := range row {
			if i >= len(widths) {
				break
			}
			if i > 0 {
				line += "  "
			}
			line += padRight(col, widths[i])
		}
		fmt.Println(line)
	}
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
