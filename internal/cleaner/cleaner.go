package cleaner

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/terratensor/text2glove/pkg/utils"
	"golang.org/x/text/unicode/norm"
)

type CleanMode string

const (
	ModeModern      CleanMode = "modern"
	ModeOldSlavonic CleanMode = "old_slavonic"
	ModeAll         CleanMode = "all"
)

type TextCleaner struct {
	re           *regexp.Regexp
	mode         CleanMode
	longWordsLog *os.File
	logEnabled   bool
	logMutex     sync.Mutex
	config       utils.Config
}

func NewWithLogger(mode CleanMode, config utils.Config) (*TextCleaner, error) {
	var logFile *os.File
	var err error

	if config.Logger.Enabled && config.Logger.LongWordsLog != "" {
		logFile, err = os.OpenFile(config.Logger.LongWordsLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
	}

	return &TextCleaner{
		re:           createCleanupRegexp(mode, config.Preserve.Dates, config.Preserve.Fractions),
		mode:         mode,
		longWordsLog: logFile,
		logEnabled:   config.Logger.Enabled,
		config:       config,
	}, nil
}

func (c *TextCleaner) Clean(text string) string {
	// Нормализация Unicode и приведение к нижнему регистру
	text = norm.NFKC.String(text)
	text = strings.ToLower(text)

	// Языковые преобразования (русский)
	if c.config.Cleaner.ReplaceYo {
		text = strings.ReplaceAll(text, "ё", "е")
	}

	// Основная очистка с учетом сохранения чисел и дат
	preserveNumbers := c.config.Preserve.Fractions || c.config.Preserve.Decimals
	c.re = createCleanupRegexp(c.mode, c.config.Preserve.Dates, preserveNumbers)
	text = c.re.ReplaceAllString(text, "")

	// Обработка дат после основной очистки
	if c.config.Preserve.Dates {
		text = processDates(text)
	}

	// Обработка чисел после основной очистки
	if c.config.Preserve.Fractions {
		text = processFractions(text)
	}

	if c.config.Preserve.Decimals {
		text = processDecimals(text)
	}

	// Замена специальных сущностей
	// text = regexp.MustCompile(`(https?://|www\.)[^\s]+`).ReplaceAllString(text, " <url> ")
	// text = regexp.MustCompile(`\S+@\S+\.\S+`).ReplaceAllString(text, " <email> ")
	// text = regexp.MustCompile(`\d+`).ReplaceAllString(text, " <num> ")
	// text = regexp.MustCompile(`#[a-zа-яё0-9_]+`).ReplaceAllString(text, " <hashtag> ")
	// text = regexp.MustCompile(`@[a-zа-яё0-9_]+`).ReplaceAllString(text, " <mention> ")

	// Удаление слишком длинных слов с логированием
	if c.logEnabled {
		text = c.removeLongWordsWithLog(text)
	} else {
		text = regexp.MustCompile(`\b\w{30,}\b`).ReplaceAllString(text, "")
	}

	// Нормализуем пробелы
	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}

func createCleanupRegexp(mode CleanMode, preserveDates, preserveNumbers bool) *regexp.Regexp {
	var pattern string

	// Базовые разрешенные символы
	baseChars := `\p{L}\p{N}\s`

	// Добавляем дополнительные разрешенные символы в зависимости от режима
	switch mode {
	case ModeModern:
		pattern = `[^` + baseChars + `а-яёa-zà-ÿğüşıöç'`
	case ModeOldSlavonic:
		oldSlavonicChars := "ѣѢѵѴіІѳѲѫѪѭѬѧѦѩѨѯѮѱѰѡѠѿѾҌҍꙋꙊꙗꙖꙙꙘꙜꙛꙝꙞꙟꙠꙡꙢꙣꙤꙥꙦꙧꙨꙩꙪꙫꙬꙭꙮѻѺѹѸѷѶѵѴѳѲѱѰѯѮѭѬѫѪѩѨѧѦѥѤѣѢѣѢѡѠџЏѾѽѼѻѺѹѸ"
		pattern = `[^` + baseChars + oldSlavonicChars
	case ModeAll:
		pattern = `[^` + baseChars
	}

	// Если нужно сохранять даты/числа, добавляем необходимые символы
	if preserveDates || preserveNumbers {
		pattern += `/\-\.\,`
	}

	// Закрываем группу символов
	pattern += `]`

	return regexp.MustCompile(pattern)
}

func (c *TextCleaner) removeLongWordsWithLog(text string) string {
	re := regexp.MustCompile(`\b(\w{30,})\b`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		c.logMutex.Lock()
		defer c.logMutex.Unlock()

		if _, err := c.longWordsLog.WriteString(match + "\n"); err != nil {
			fmt.Printf("Failed to log long word: %v\n", err)
		}
		return ""
	})
}

func (c *TextCleaner) Close() error {
	if c.longWordsLog != nil {
		return c.longWordsLog.Close()
	}
	return nil
}

func processDates(text string) string {
	// Объединяем все форматы дат в одно регулярное выражение
	datePatterns := []string{
		`\b\d{4}-\d{2}-\d{2}\b`,         // ISO (2000-12-31)
		`\b\d{1,2}\.\d{1,2}\.\d{2,4}\b`, // 31.12.2000
		`\b\d{1,2}/\d{1,2}/\d{2,4}\b`,   // 12/31/2000
		`\b\d{1,2}-\d{1,2}-\d{2,4}\b`,   // 12-31-2000
	}

	for _, pattern := range datePatterns {
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllStringFunc(text, func(match string) string {
			return " " + match + " " // Добавляем пробелы вокруг даты
		})
	}
	return text
}

func processFractions(text string) string {
	re := regexp.MustCompile(`\b\d+/\d+\b`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		return " " + match + " "
	})
}

func processDecimals(text string) string {
	re := regexp.MustCompile(`\b\d+[\.,]\d+\b`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		// Нормализуем разделитель к точке
		normalized := strings.Replace(match, ",", ".", 1)
		return " " + normalized + " "
	})
}
