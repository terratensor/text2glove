package writer

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

type ResultWriter struct {
	filePath   string
	bufferSize int
}

func New(filePath string, bufferSize int) *ResultWriter {
	return &ResultWriter{
		filePath:   filePath,
		bufferSize: bufferSize,
	}
}

func (w *ResultWriter) Write(textChan <-chan string) {
	file, err := os.Create(w.filePath)
	if err != nil {
		fmt.Printf("Failed to create output file: %v\n", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, w.bufferSize)
	defer writer.Flush()

	var totalLines, totalBytes uint64
	startTime := time.Now()
	lastReport := startTime

	for text := range textChan {
		n, err := writer.WriteString(text + "\n")
		if err != nil {
			fmt.Printf("Write error: %v\n", err)
			continue
		}

		totalLines++
		totalBytes += uint64(n)

		if time.Since(lastReport) > 5*time.Second {
			rate := float64(totalBytes) / time.Since(startTime).Seconds() / 1024
			fmt.Printf("Written: %d lines (%.1f KB/s)\n", totalLines, rate)
			lastReport = time.Now()
		}
	}

	fmt.Printf("Finished writing: %d lines, %d bytes\n", totalLines, totalBytes)
}
