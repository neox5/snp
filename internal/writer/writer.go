package writer

import (
	"bufio"
	"io"
)

// LineTracker tracks the current line number while writing
type LineTracker struct {
	w           *bufio.Writer
	currentLine int
}

// NewLineTracker creates a new line tracking writer
func NewLineTracker(w io.Writer) *LineTracker {
	return &LineTracker{
		w:           bufio.NewWriter(w),
		currentLine: 1,
	}
}

// WriteLine writes a line and increments line counter
func (lt *LineTracker) WriteLine(s string) error {
	if _, err := lt.w.WriteString(s + "\n"); err != nil {
		return err
	}
	lt.currentLine++
	return nil
}

// WriteString writes without newline or tracking
func (lt *LineTracker) WriteString(s string) error {
	_, err := lt.w.WriteString(s)
	return err
}

// CurrentLine returns current line number (1-based)
func (lt *LineTracker) CurrentLine() int {
	return lt.currentLine
}

// Flush flushes the underlying buffer
func (lt *LineTracker) Flush() error {
	return lt.w.Flush()
}
