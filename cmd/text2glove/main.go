package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/terratensor/text2glove/internal/cleaner"
	"github.com/terratensor/text2glove/internal/lemmatizer"
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
	pflag.String("cleaner_mode", "unicode_letters_and_numbers", "Cleaner mode: modern|old_slavonic|all|unicode_letters")
	pflag.Bool("normalize", true, "Apply Unicode normalization")
	pflag.Bool("lemmatize", false, "Enable lemmatization with mystem")
	pflag.String("mystem_path", "", "Path to mystem binary (default: look in PATH)")
	pflag.String("mystem_flags", "-ld", "Mystem flags")
}

func main() {
	pflag.Parse()

	// 1. Инициализация Viper с явными значениями по умолчанию
	v := viper.New()
	v.SetDefault("lemmatization.enable", false)
	v.SetDefault("lemmatization.mystem_path", "")
	v.SetDefault("lemmatization.mystem_flags", "-ld")

	// 2. Привязка флагов командной строки (высший приоритет)
	if err := v.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatalf("Failed to bind flags: %v", err)
	}

	// 3. Загрузка конфига если указан
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}
	}

	// 4. Сборка финальной конфигурации
	config := utils.Config{
		InputDir:     v.GetString("input"),
		OutputFile:   v.GetString("output"),
		WorkersCount: v.GetInt("workers"),
		BufferSize:   v.GetInt("buffer_size"),
		ReportEvery:  v.GetInt("report_every"),
	}
	config.Cleaner.Mode = v.GetString("cleaner_mode")
	config.Cleaner.Normalize = v.GetBool("normalize")
	config.Lemmatization.Enable = v.GetBool("lemmatize") || v.GetBool("lemmatization.enable")
	config.Lemmatization.MystemPath = v.GetString("mystem_path")
	if config.Lemmatization.MystemPath == "" {
		config.Lemmatization.MystemPath = v.GetString("lemmatization.mystem_path")
	}
	config.Lemmatization.MystemFlags = v.GetString("mystem_flags")
	if config.Lemmatization.MystemFlags == "" {
		config.Lemmatization.MystemFlags = v.GetString("lemmatization.mystem_flags")
	}

	// 5. Автопоиск mystem если путь не указан
	if config.Lemmatization.Enable && config.Lemmatization.MystemPath == "" {
		if path, err := exec.LookPath("mystem"); err == nil {
			config.Lemmatization.MystemPath = path
		} else {
			log.Fatal("Mystem not found in PATH. Please specify --mystem_path")
		}
	}

	// Добавляем чтение настроек логгера
	config.Logger.Enabled = v.GetBool("logger.enabled")
	config.Logger.LongWordsLog = v.GetString("logger.long_words_log")

	startPipeline(config)
}

func startPipeline(config utils.Config) {
	startTime := time.Now()

	fmt.Println("=== Starting Text2Glove ===")
	if viper.ConfigFileUsed() != "" {
		fmt.Printf("Config file: %s\n", viper.ConfigFileUsed())
	}
	fmt.Printf("Input directory: %s\n", config.InputDir)
	fmt.Printf("Output file: %s\n", config.OutputFile)
	fmt.Printf("Number of workers: %v\n", config.WorkersCount)
	fmt.Printf("Cleaner mode: %s\n", config.Cleaner.Mode)
	fmt.Printf("Unicode normalization: %v\n", config.Cleaner.Normalize)
	fmt.Printf("Lemmatization enabled: %v\n", config.Lemmatization.Enable)
	fmt.Printf("Logger enabled: %v\n", config.Logger.Enabled)
	fmt.Printf("Long words log: %v\n", config.Logger.LongWordsLog)
	if config.Lemmatization.Enable {
		fmt.Printf("Mystem path: %s\n", config.Lemmatization.MystemPath)
		fmt.Printf("Mystem flags: %s\n", config.Lemmatization.MystemFlags)
	}

	// Инициализация cleaner с опциями
	cleanOptions := cleaner.CleanOptions{
		KeepNumbers:      config.Cleaner.KeepNumbers,      // из конфига
		KeepRomanNumbers: config.Cleaner.KeepRomanNumbers, // из конфига
	}

	textCleaner := cleaner.New(
		cleaner.CleanMode(config.Cleaner.Mode),
		cleanOptions,
	)

	// Инициализация лемматизатора с логгером
	var lem *lemmatizer.Lemmatizer
	var err error
	if config.Lemmatization.Enable {
		lem, err = lemmatizer.New(
			config.Lemmatization.MystemPath,
			config.Lemmatization.MystemFlags,
			config.Logger.Enabled,      // передаем флаг включения логирования
			config.Logger.LongWordsLog, // передаем путь к лог-файлу
		)
		if err != nil {
			log.Fatalf("Failed to initialize lemmatizer: %v", err)
		}
		defer lem.Close()
	}

	fileProcessor := processor.New(textCleaner, lem, config.Lemmatization.Enable)
	resultWriter := writer.New(config.OutputFile, config.BufferSize)

	// Обработка файлов
	if err := processFiles(config, fileProcessor, resultWriter); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n=== Processing completed in %v ===\n", time.Since(startTime))
}

func processFiles(config utils.Config, processor *processor.FileProcessor, resultWriter *writer.ResultWriter) error {
	// Исправленный поиск файлов с пробелами в именах
	pattern := filepath.Join(config.InputDir, "*.gz")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to list files: %v", err)
	}

	// Альтернативный метод для имен с пробелами
	if len(matches) == 0 {
		files, err := filepath.Glob(filepath.Join(config.InputDir, "*"))
		if err != nil {
			return fmt.Errorf("failed to list files: %v", err)
		}
		for _, f := range files {
			if strings.HasSuffix(strings.ToLower(f), ".gz") {
				matches = append(matches, f)
			}
		}
	}

	totalFiles := len(matches)
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
				printProgress(totalProcessed.Load(), uint64(totalFiles), resultWriter)
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
			processor.Work(id, fileChan, textChan, progressChan, resultWriter)
		}(i + 1)
	}

	// Отправляем файлы в канал для обработки
	go func() {
		for _, file := range matches {
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
	fmt.Printf("  Time:      %v\n", stats.Duration.Round(time.Second))
	fmt.Printf("  Lines:     %d\n", stats.Lines)
	fmt.Printf("  Corrupted: %d\n", stats.Corrupted) // Новая статистика
	fmt.Printf("  Data:      %.1f MB\n", mb)
	fmt.Printf("  Speed:     %.1f KB/s\n", speed)
}
