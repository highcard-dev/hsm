package utils

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// ProgressWriter wraps an io.Writer to display download progress
type ProgressWriter struct {
	Total      int64
	written    int64
	lastUpdate time.Time
	startTime  time.Time
}

// NewProgressWriter creates a new progress writer with the given total size
func NewProgressWriter(total int64) *ProgressWriter {
	return &ProgressWriter{
		Total:     total,
		startTime: time.Now(),
	}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.written += int64(n)

	if time.Since(pw.lastUpdate) > 100*time.Millisecond {
		pw.print()
		pw.lastUpdate = time.Now()
	}
	return n, nil
}

func (pw *ProgressWriter) print() {
	const barWidth = 40

	percent := float64(pw.written) / float64(pw.Total)
	if pw.Total <= 0 {
		percent = 0
	}

	filled := int(percent * barWidth)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	elapsed := time.Since(pw.startTime).Seconds()
	speed := float64(pw.written) / elapsed / 1024 / 1024

	fmt.Printf("\r  [%s] %.1f%% • %.1f/%.1f MB • %.2f MB/s",
		bar, percent*100,
		float64(pw.written)/1024/1024,
		float64(pw.Total)/1024/1024,
		speed,
	)
}

// Finish prints final progress and newline
func (pw *ProgressWriter) Finish() {
	pw.print()
	fmt.Println()
}

// WrapReader returns an io.Reader that tracks progress while reading
func (pw *ProgressWriter) WrapReader(r io.Reader) io.Reader {
	return io.TeeReader(r, pw)
}
