package cleaner

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

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
}

func New(mode CleanMode) *TextCleaner {
	return &TextCleaner{
		re:   createCleanupRegexp(mode),
		mode: mode,
	}
}

func NewWithLogger(mode CleanMode, logPath string, enabled bool) (*TextCleaner, error) {
	var logFile *os.File
	var err error

	if enabled && logPath != "" {
		logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
	}

	return &TextCleaner{
		re:           createCleanupRegexp(mode),
		mode:         mode,
		longWordsLog: logFile,
		logEnabled:   enabled,
	}, nil
}

func (c *TextCleaner) Clean(text string) string {
	// Нормализуем Unicode (NFKC - совмещает совместимые символы)
	text = norm.NFKC.String(text)

	// Приводим к нижнему регистру с учетом Unicode
	text = strings.ToLower(text)

	// Замена специальных сущностей
	text = regexp.MustCompile(`(https?://|www\.)[^\s]+`).ReplaceAllString(text, " <url> ")
	text = regexp.MustCompile(`\S+@\S+\.\S+`).ReplaceAllString(text, " <email> ")
	text = regexp.MustCompile(`\d+`).ReplaceAllString(text, " <num> ")
	text = regexp.MustCompile(`#[a-zа-яё0-9_]+`).ReplaceAllString(text, " <hashtag> ")
	text = regexp.MustCompile(`@[a-zа-яё0-9_]+`).ReplaceAllString(text, " <mention> ")

	// Основная очистка текста
	text = c.re.ReplaceAllString(text, "")

	// Нормализация пунктуации
	// text = regexp.MustCompile(`(!|\?|\.){2,}`).ReplaceAllString(text, "$1$1")

	// Языковые преобразования (русский)
	text = strings.ReplaceAll(text, "ё", "е")

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
		// Разрешаем < и > для токенов, добавляем апостроф для английского
		pattern = `[^\p{L}\p{N}\sа-яёa-zà-ÿğüşıöç'<>]`
	case ModeOldSlavonic:
		oldSlavonicChars := "ѣѢѵѴіІѳѲѫѪѭѬѧѦѩѨѯѮѱѰѡѠѿѾҌҍꙋꙊꙗꙖꙙꙘꙜꙛꙝꙞꙟꙠꙡꙢꙣꙤꙥꙦꙧꙨꙩꙪꙫꙬꙭꙮѻѺѹѸѷѶѵѴѳѲѱѰѯѮѭѬѫѪѩѨѧѦѥѤѣѢѣѢѡѠџЏѾѽѼѻѺѹѸ"
		pattern = `[^\p{L}\p{N}\s` + oldSlavonicChars + `<>]`
	case ModeAll:
		// Все буквы Unicode + <>
		pattern = `[^\p{L}\p{N}\s<>]`
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
