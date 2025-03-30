package processor

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"strings"

	"github.com/terratensor/text2glove/internal/cleaner"
)

const (
	maxTokenSize = 10 * 1024 * 1024 // 10MB
)

type FileProcessor struct {
	cleaner cleaner.TextCleaner // используем интерфейс вместо конкретной реализации
}

func New(cleaner cleaner.TextCleaner) *FileProcessor {
	return &FileProcessor{
		cleaner: cleaner,
	}
}

func (p *FileProcessor) Work(id int, fileChan <-chan string, textChan chan<- string, progressChan chan<- int) {
	var processed int

	for file := range fileChan {
		text, err := p.processFile(file)
		if err != nil {
			fmt.Printf("\r\x1b[31mError:\x1b[0m %s: %v\n", file, err)
			continue
		}

		if text != "" {
			textChan <- text
		}

		processed++
		if processed%100 == 0 {
			progressChan <- processed
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

	buf := make([]byte, 0, maxTokenSize)
	scanner.Buffer(buf, maxTokenSize)

	for scanner.Scan() {
		line := scanner.Text()
		cleanLine := p.cleaner.Clean(line)
		if cleanLine != "" {
			builder.WriteString(cleanLine)
			builder.WriteRune(' ')
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanner error: %v", err)
	}

	return builder.String(), nil
}
