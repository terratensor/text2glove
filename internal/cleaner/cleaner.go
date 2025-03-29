package cleaner

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type CleanMode string

const (
	ModeModern      CleanMode = "modern"
	ModeOldSlavonic CleanMode = "old_slavonic"
	ModeAll         CleanMode = "all"
)

type TextCleaner struct {
	re   *regexp.Regexp
	mode CleanMode
}

func New(mode CleanMode) *TextCleaner {
	return &TextCleaner{
		re:   createCleanupRegexp(mode),
		mode: mode,
	}
}

func (c *TextCleaner) Clean(text string) string {
	// Нормализуем Unicode (NFKC - совмещает совместимые символы)
	text = norm.NFKC.String(text)

	// Приводим к нижнему регистру с учетом Unicode
	text = strings.ToLower(text)

	// Удаляем нежелательные символы
	text = c.re.ReplaceAllString(text, "")

	// Нормализуем пробелы
	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}

func createCleanupRegexp(mode CleanMode) *regexp.Regexp {
	var pattern string

	switch mode {
	case ModeModern:
		// Современные языки (русский, английский, европейские)
		pattern = `[^\p{L}\p{N}\sа-яёa-zà-ÿğüşıöç]`
	case ModeOldSlavonic:
		// Старославянские символы + современные
		pattern = `[^\p{L}\p{N}\s\p{In_Cyrillic}\p{In_Cyrillic_Supplement}\p{In_Cyrillic_Extended-A}\p{In_Cyrillic_Extended-B}\p{In_Cyrillic_Extended-C}]`
	case ModeAll:
		// Все буквы Unicode
		pattern = `[^\p{L}\p{N}\s]`
	default:
		pattern = `[^\p{L}\p{N}\s]`
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		panic("Failed to compile regexp: " + err.Error())
	}
	return re
}

// Альтернативный метод для точной обработки символов
func (c *TextCleaner) CleanManual(text string) string {
	var builder strings.Builder
	builder.Grow(len(text))

	for _, r := range text {
		// Проверяем категорию Unicode
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder.WriteRune(unicode.ToLower(r))
		} else if unicode.IsSpace(r) {
			builder.WriteRune(' ')
		}
		// Специальная обработка для старославянских символов
		// может быть добавлена здесь при необходимости
	}

	// Нормализуем пробелы
	result := strings.Join(strings.Fields(builder.String()), " ")
	return strings.TrimSpace(result)
}
