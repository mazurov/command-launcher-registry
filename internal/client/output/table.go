package output

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// TableWriter wraps tabwriter for formatted output
type TableWriter struct {
	writer *tabwriter.Writer
}

// NewTableWriter creates a new table writer
func NewTableWriter() *TableWriter {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	return &TableWriter{writer: w}
}

// WriteHeader writes table headers
func (t *TableWriter) WriteHeader(headers ...string) {
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(t.writer, "\t")
		}
		fmt.Fprint(t.writer, h)
	}
	fmt.Fprintln(t.writer)
}

// WriteRow writes a table row
func (t *TableWriter) WriteRow(values ...string) {
	for i, v := range values {
		if i > 0 {
			fmt.Fprint(t.writer, "\t")
		}
		fmt.Fprint(t.writer, v)
	}
	fmt.Fprintln(t.writer)
}

// Flush writes buffered output
func (t *TableWriter) Flush() error {
	return t.writer.Flush()
}

// PrintSuccess prints a success message with checkmark
func PrintSuccess(message string) {
	fmt.Printf("✓ %s\n", message)
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Fprintf(os.Stderr, "✗ %s\n", message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Fprintf(os.Stderr, "⚠ %s\n", message)
}
