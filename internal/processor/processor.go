package processor

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"strings"

	"github.com/terratensor/text2glove/internal/cleaner"
	"github.com/terratensor/text2glove/internal/detector"
	"github.com/terratensor/text2glove/internal/writer"
)

const (
	maxTokenSize = 10 * 1024 * 1024 // 10MB
)

type FileProcessor struct {
	cleaner *cleaner.TextCleaner
}

func New(cleaner *cleaner.TextCleaner) *FileProcessor {
	return &FileProcessor{
		cleaner: cleaner,
	}
}

func (p *FileProcessor) Work(id int, fileChan <-chan string, textChan chan<- string, progressChan chan<- int, resultWriter *writer.ResultWriter) {
	var processed, corrupted int

	for file := range fileChan {
		text, err := p.processFile(file)
		if err != nil {
			fmt.Printf("\r\x1b[31mError:\x1b[0m %s: %v\n", file, err)
			continue
		}

		if text != "" {
			if detector.IsCorrupted(text) {
				corrupted++
				resultWriter.IncrementCorrupted()
				continue // Пропускаем битые тексты
			}
			textChan <- text
		}

		processed++
		if processed%100 == 0 {
			progressChan <- processed
			if corrupted > 0 {
				fmt.Printf("\n\x1b[33mWorker %d: detected %d corrupted files\x1b[0m\n", id, corrupted)
				corrupted = 0
			}
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
