package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"sync"
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
	pflag.String("cleaner_mode", "old_slavonic", "Cleaner mode: modern|old_slavonic|all")
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

func processFiles(config utils.Config, processor *processor.FileProcessor, writer *writer.ResultWriter) error {
	files, err := filepath.Glob(filepath.Join(config.InputDir, "*.gz"))
	if err != nil {
		return fmt.Errorf("failed to list files: %v", err)
	}

	totalFiles := len(files)
	fmt.Printf("Found %d files to process\n", totalFiles)

	var wg sync.WaitGroup
	fileChan := make(chan string, config.WorkersCount*2)
	textChan := make(chan string, config.WorkersCount*2)

	// Start writer
	go writer.Write(textChan)

	// Start workers
	for i := 0; i < config.WorkersCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			processor.Work(id, fileChan, textChan, config.ReportEvery, totalFiles)
		}(i + 1)
	}

	// Feed files to workers
	go func() {
		for _, file := range files {
			fileChan <- file
		}
		close(fileChan)
	}()

	wg.Wait()
	close(textChan)

	return nil
}
