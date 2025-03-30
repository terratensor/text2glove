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
		re:           createCleanupRegexp(mode),
		mode:         mode,
		longWordsLog: logFile,
		logEnabled:   config.Logger.Enabled,
		config:       config,
	}, nil
}

func (c *TextCleaner) Clean(text string) string {
	// Нормализуем Unicode (NFKC - совмещает совместимые символы)
	text = norm.NFKC.String(text)

	// Приводим к нижнему регистру с учетом Unicode
	text = strings.ToLower(text)

	// Языковые преобразования (русский)
	if c.config.Cleaner.ReplaceYo {
		text = strings.ReplaceAll(text, "ё", "е")
	}

	// Сохранение дат
	if c.config.Preserve.Dates {
		// ISO (2000-12-31), американский (12/31/2000), европейский (31.12.2000)
		datePattern := `(?:\d{4}-\d{2}-\d{2})|(?:\d{1,2}[/\-\.]\d{1,2}[/\-\.]\d{2,4})`
		text = regexp.MustCompile(datePattern).ReplaceAllStringFunc(text, func(match string) string {
			return " " + match + " " // Добавляем пробелы для сохранения как отдельного токена
		})
	}

	// Сохранение дробных чисел
	if c.config.Preserve.Fractions {
		// Простые дроби (2/3, 1/4 и т.д.)
		fractionPattern := `\b\d+/\d+\b`
		text = regexp.MustCompile(fractionPattern).ReplaceAllStringFunc(text, func(match string) string {
			return " " + match + " "
		})
	}

	// Сохранение десятичных чисел
	if c.config.Preserve.Decimals {
		// Числа с плавающей точкой (5.20, 5,20)
		decimalPattern := `\b\d+[\.,]\d+\b`
		text = regexp.MustCompile(decimalPattern).ReplaceAllStringFunc(text, func(match string) string {
			// Нормализуем разделитель к точке
			normalized := strings.Replace(match, ",", ".", 1)
			return " " + normalized + " "
		})
	}

	// Замена специальных сущностей
	// text = regexp.MustCompile(`(https?://|www\.)[^\s]+`).ReplaceAllString(text, " <url> ")
	// text = regexp.MustCompile(`\S+@\S+\.\S+`).ReplaceAllString(text, " <email> ")
	// text = regexp.MustCompile(`\d+`).ReplaceAllString(text, " <num> ")
	// text = regexp.MustCompile(`#[a-zа-яё0-9_]+`).ReplaceAllString(text, " <hashtag> ")
	// text = regexp.MustCompile(`@[a-zа-яё0-9_]+`).ReplaceAllString(text, " <mention> ")

	// Основная очистка текста
	text = c.re.ReplaceAllString(text, "")

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

func createCleanupRegexp(mode CleanMode) *regexp.Regexp {
	var pattern string

	switch mode {
	case ModeModern:
		// Современные языки: русский, английский, основные европейские
		pattern = `[^\p{L}\p{N}\sа-яёa-zà-ÿğüşıöç]`

	case ModeOldSlavonic:
		// Старославянские символы (явное перечисление)
		oldSlavonicChars := "ѣѢѵѴіІѳѲѫѪѭѬѧѦѩѨѯѮѱѰѡѠѿѾҌҍꙋꙊꙗꙖꙙꙘꙜꙛꙝꙞꙟꙠꙡꙢꙣꙤꙥꙦꙧꙨꙩꙪꙫꙬꙭꙮѻѺѹѸѷѶѵѴѳѲѱѰѯѮѭѬѫѪѩѨѧѦѥѤѣѢѣѢѡѠџЏѾѽѼѻѺѹѸ"
		pattern = `[^\p{L}\p{N}\s` + oldSlavonicChars + `]`

	case ModeAll:
		// Все буквы Unicode
		pattern = `[^\p{L}\p{N}\s]`
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		panic("Failed to compile regexp for mode " + string(mode) + ": " + err.Error())
	}
	return re
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

func escapeToken(token string) string {
	return " " + token + " " // Добавляем пробелы для отделения токена
}
