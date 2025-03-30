package writer

import (
	"bufio"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

type Stats struct {
	Lines    uint64
	Bytes    uint64
	Duration time.Duration
}

type ResultWriter struct {
	filePath   string
	bufferSize int
	totalLines atomic.Uint64
	totalBytes atomic.Uint64
	startTime  time.Time
}

func New(filePath string, bufferSize int) *ResultWriter {
	return &ResultWriter{
		filePath:   filePath,
		bufferSize: bufferSize,
		startTime:  time.Now(),
	}
}

func (w *ResultWriter) Write(textChan <-chan string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\x1b[31mWriter panic: %v\x1b[0m\n", r)
		}
	}()

	file, err := os.Create(w.filePath)
	if err != nil {
		fmt.Printf("\x1b[31mFailed to create output file: %v\x1b[0m\n", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, w.bufferSize)
	defer writer.Flush()

	for text := range textChan {
		if text == "" {
			continue
		}
		_, err := writer.WriteString(text + "\n")
		if err != nil {
			fmt.Printf("\x1b[31mWrite error: %v\x1b[0m\n", err)
			continue
		}
		w.totalLines.Add(1)
		w.totalBytes.Add(uint64(len(text) + 1)) // +1 for newline
	}
}

func (w *ResultWriter) GetStats() Stats {
	return Stats{
		Lines:    w.totalLines.Load(),
		Bytes:    w.totalBytes.Load(),
		Duration: time.Since(w.startTime),
	}
}
