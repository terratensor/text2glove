package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/terratensor/text2glove/internal/cleaner"
	"github.com/terratensor/text2glove/internal/processor"
	"github.com/terratensor/text2glove/internal/writer"
	"github.com/terratensor/text2glove/pkg/utils"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	configFile string
)

func init() {
	pflag.StringVarP(&configFile, "config", "c", "", "Path to config file")
	pflag.String("input", "./data", "Input directory with .gz files")
	pflag.String("output", "./output.txt", "Output file path")
	pflag.Int("workers", runtime.NumCPU(), "Number of workers")
	pflag.Int("buffer_size", 1024*1024, "Writer buffer size in bytes")
	pflag.Int("report_every", 100, "Report progress every N files")
	pflag.String("cleaner_mode", "all", "Cleaner mode: modern|old_slavonic|all")
	pflag.Bool("normalize", true, "Apply Unicode normalization")
}

func main() {
	pflag.Parse()

	// Load configuration
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatalf("Failed to bind flags: %v", err)
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Failed to read config: %v", err)
		}
	}

	startTime := time.Now()

	config := utils.Config{
		InputDir:     viper.GetString("input"),
		OutputFile:   viper.GetString("output"),
		WorkersCount: viper.GetInt("workers"),
		BufferSize:   viper.GetInt("buffer_size"),
		ReportEvery:  viper.GetInt("report_every"),
	}
	config.Cleaner.Mode = viper.GetString("cleaner_mode")
	config.Cleaner.Normalize = viper.GetBool("normalize")

	fmt.Println("=== Starting Text2Glove ===")
	fmt.Printf("Number of workers: %v\n", config.WorkersCount)
	fmt.Printf("Cleaner mode: %s\n", config.Cleaner.Mode)
	fmt.Printf("Unicode normalization: %v\n", config.Cleaner.Normalize)

	// Initialize components
	textCleaner := cleaner.New(cleaner.CleanMode(config.Cleaner.Mode))
	fileProcessor := processor.New(textCleaner)
	resultWriter := writer.New(config.OutputFile, config.BufferSize)

	// Start processing
	if err := processFiles(config, fileProcessor, resultWriter); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n=== Processing completed in %v ===\n", time.Since(startTime))
}

func processFiles(config utils.Config, processor *processor.FileProcessor, resultWriter *writer.ResultWriter) error {
	files, err := filepath.Glob(filepath.Join(config.InputDir, "*.gz"))
	if err != nil {
		return fmt.Errorf("failed to list files: %v", err)
	}

	totalFiles := uint64(len(files))
	if totalFiles == 0 {
		return fmt.Errorf("no .gz files found in directory %s", config.InputDir)
	}
	fmt.Printf("Found %d files to process\n", totalFiles)

	// Каналы для работы
	fileChan := make(chan string, config.WorkersCount*2)
	textChan := make(chan string, config.WorkersCount*2)
	progressChan := make(chan int, config.WorkersCount)
	done := make(chan struct{})

	// Группа ожидания для рабочих
	var wg sync.WaitGroup

	// Запускаем писателя в отдельной горутине
	go func() {
		resultWriter.Write(textChan)
		close(done)
	}()

	// Горутина для вывода прогресса
	go func() {
		var totalProcessed atomic.Uint64
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case n, ok := <-progressChan:
				if !ok {
					return
				}
				totalProcessed.Add(uint64(n))
			case <-ticker.C:
				printProgress(totalProcessed.Load(), totalFiles, resultWriter)
			case <-done:
				return
			}
		}
	}()

	// Запускаем рабочих
	for i := 0; i < config.WorkersCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			processor.Work(id, fileChan, textChan, progressChan)
		}(i + 1)
	}

	// Отправляем файлы в канал для обработки
	go func() {
		for _, file := range files {
			fileChan <- file
		}
		close(fileChan)
	}()

	wg.Wait()
	close(textChan)
	close(progressChan)

	// Ждем завершения писателя
	<-done

	// Вывод финальной статистики
	printFinalStats(resultWriter)

	return nil
}

func printProgress(processed, total uint64, writer *writer.ResultWriter) {
	width := 50
	percent := float64(processed) / float64(total)

	// Защита от переполнения и отрицательных значений
	if percent > 1.0 {
		percent = 1.0
	} else if percent < 0 {
		percent = 0
	}

	filled := int(float64(width) * percent)
	if filled > width {
		filled = width
	}

	stats := writer.GetStats()
	var speed float64
	if stats.Duration.Seconds() > 0 {
		speed = float64(stats.Bytes) / 1024 / stats.Duration.Seconds()
	}

	bar := strings.Repeat("=", filled) + strings.Repeat(" ", width-filled)
	fmt.Printf("\r\x1b[36mProcessing:\x1b[0m [%s] %6.2f%% | \x1b[33mSpeed:\x1b[0m %7.1f KB/s | \x1b[32mLines:\x1b[0m %d",
		bar, percent*100, speed, stats.Lines)
}

func printFinalStats(writer *writer.ResultWriter) {
	stats := writer.GetStats()
	speed := float64(stats.Bytes) / 1024 / stats.Duration.Seconds()
	mb := float64(stats.Bytes) / 1024 / 1024

	fmt.Printf("\n\n\x1b[1m=== Processing completed ===\x1b[0m\n")
	fmt.Printf("  Time:    %v\n", stats.Duration.Round(time.Second))
	fmt.Printf("  Lines:   %d\n", stats.Lines)
	fmt.Printf("  Data:    %.1f MB\n", mb)
	fmt.Printf("  Speed:   %.1f KB/s\n", speed)
}
