package processor

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"github.com/terratensor/text2glove/internal/cleaner"
)

const (
	maxTokenSize = 10 * 1024 * 1024 // 10MB максимальный размер токена
)

type FileProcessor struct {
	cleaner *cleaner.TextCleaner
}

func New(cleaner *cleaner.TextCleaner) *FileProcessor {
	return &FileProcessor{
		cleaner: cleaner,
	}
}

func (p *FileProcessor) Work(id int, fileChan <-chan string, textChan chan<- string, reportEvery, totalFiles int) {
	var processed uint64

	for file := range fileChan {
		text, err := p.processFile(file)
		if err != nil {
			fmt.Printf("Worker %d: error processing %s: %v\n", id, file, err)
			continue
		}

		textChan <- text

		// Report progress
		newProcessed := atomic.AddUint64(&processed, 1)
		if newProcessed%uint64(reportEvery) == 0 {
			fmt.Printf("Worker %d: processed %d/%d (%.1f%%)\n",
				id, newProcessed, totalFiles, float64(newProcessed)/float64(totalFiles)*100)
		}
	}
}

func (p *FileProcessor) processFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("gzip error: %v", err)
	}
	defer gz.Close()

	var builder strings.Builder
	scanner := bufio.NewScanner(gz)

	// Увеличиваем буфер сканера
	buf := make([]byte, 0, maxTokenSize)
	scanner.Buffer(buf, maxTokenSize)

	for scanner.Scan() {
		line := scanner.Text()
		cleanLine := p.cleaner.Clean(line)
		if cleanLine != "" {
			builder.WriteString(cleanLine)
			builder.WriteString(" ")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanner error: %v", err)
	}

	return builder.String(), nil
}
