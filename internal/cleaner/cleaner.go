package cleaner

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

type CleanMode string

const (
	ModeModern                   CleanMode = "modern"
	ModeOldSlavonic              CleanMode = "old_slavonic"
	ModeAll                      CleanMode = "all"
	ModeUnicodeLettersAndNumbers CleanMode = "unicode_letters_and_numbers"
)

type CleanOptions struct {
	KeepNumbers      bool // сохранять арабские цифры
	KeepRomanNumbers bool // сохранять римские цифры
}

type TextCleaner struct {
	re           *regexp.Regexp
	urlRe        *regexp.Regexp
	emailRe      *regexp.Regexp
	romanNumRe   *regexp.Regexp
	numbersRe    *regexp.Regexp
	whitespaceRe *regexp.Regexp
	mode         CleanMode
	options      CleanOptions
}

func New(mode CleanMode, options CleanOptions) *TextCleaner {
	cleaner := &TextCleaner{
		urlRe:        regexp.MustCompile(`(https?://|www\.)[^\s]+`),
		emailRe:      regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		whitespaceRe: regexp.MustCompile(`\s+`),
		mode:         mode,
		options:      options,
	}

	// Инициализация regex для чисел только если нужно
	if !options.KeepRomanNumbers {
		cleaner.romanNumRe = regexp.MustCompile(`\b[IVXLCDMivxlcdm]+\b`)
	}
	if !options.KeepNumbers {
		cleaner.numbersRe = regexp.MustCompile(`\b\d+\b`)
	}

	cleaner.re = createCleanupRegexp(mode, options)
	return cleaner
}

func (c *TextCleaner) Clean(text string) string {
	// 1. Восстановление UTF-8
	text = c.fixUTF8(text)

	// 2. Удаление нулевых байтов
	text = c.removeNullBytes(text)

	// 3. Нормализация Unicode
	text = norm.NFKC.String(text)

	// 4. Замена проблемных символов
	text = c.replaceControlChars(text)
	text = c.replaceUnicodeReplacementChars(text)

	// 5. Удаление URL и email
	text = c.urlRe.ReplaceAllString(text, " ")
	text = c.emailRe.ReplaceAllString(text, " ")

	// 6. Приведение к нижнему регистру
	text = strings.ToLower(text)

	// 7. Удаление чисел (если нужно)
	if !c.options.KeepRomanNumbers && c.romanNumRe != nil {
		text = c.romanNumRe.ReplaceAllString(text, " ")
	}
	if !c.options.KeepNumbers && c.numbersRe != nil {
		text = c.numbersRe.ReplaceAllString(text, " ")
	}

	// 8. Удаление нежелательных символов по режиму
	text = c.re.ReplaceAllString(text, " ")

	// 9. Нормализация пробелов
	text = c.whitespaceRe.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// fixUTF8 заменяет битые UTF-8 последовательности на символ замены
func (c *TextCleaner) fixUTF8(text string) string {
	if !utf8.ValidString(text) {
		var buf strings.Builder
		buf.Grow(len(text))
		for i, r := range text {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(text[i:])
				if size == 1 {
					buf.WriteRune(' ')
					continue
				}
			}
			buf.WriteRune(r)
		}
		return buf.String()
	}
	return text
}

// removeNullBytes удаляет нулевые байты из строки
func (c *TextCleaner) removeNullBytes(text string) string {
	var buf strings.Builder
	buf.Grow(len(text))
	for _, r := range text {
		if r != 0 {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// replaceControlChars заменяет управляющие символы на пробелы
func (c *TextCleaner) replaceControlChars(text string) string {
	var buf strings.Builder
	buf.Grow(len(text))
	for _, r := range text {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			buf.WriteRune(' ')
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// replaceUnicodeReplacementChars заменяет символы замены Unicode (�) на пробелы
func (c *TextCleaner) replaceUnicodeReplacementChars(text string) string {
	return strings.ReplaceAll(text, "\uFFFD", " ")
}

func createCleanupRegexp(mode CleanMode, options CleanOptions) *regexp.Regexp {
	var pattern string

	switch mode {
	case ModeModern:
		// Современные языки: русский, английский, основные европейские
		if options.KeepNumbers {
			pattern = `[^\p{L}\p{N}\sа-яёa-zà-ÿğüşıöç.,!?;:'"-]`
		} else {
			pattern = `[^\p{L}\sа-яёa-zà-ÿğüşıöç.,!?;:'"-]`
		}

	case ModeOldSlavonic:
		oldSlavonicChars := "ѣѢѵѴіІѳѲѫѪѭѬѧѦѩѨѯѮѱѰѡѠѿѾҌҍꙋꙊꙗꙖꙙꙘꙜꙛꙝꙞꙟꙠꙡꙢꙣꙤꙥꙦꙧꙨꙩꙪꙫꙬꙭꙮѻѺѹѸѷѶѵѴѳѲѱѰѯѮѭѬѫѪѩѨѧѦѥѤѣѢѣѢѡѠџЏѾѽѼѻѺѹѸ"
		if options.KeepNumbers {
			pattern = `[^\p{L}\p{N}\s` + oldSlavonicChars + `.,!?;:'"-]`
		} else {
			pattern = `[^\p{L}\s` + oldSlavonicChars + `.,!?;:'"-]`
		}

	case ModeAll:
		if options.KeepNumbers {
			pattern = `[^\p{L}\p{N}\s.,!?;:'"-]`
		} else {
			pattern = `[^\p{L}\s.,!?;:'"-]`
		}

	case ModeUnicodeLettersAndNumbers:
		// Все буквы Unicode и цифры и без CJK иероглифов
		// Исключаем следующие Unicode-блоки:
		// \p{Han} - китайские иероглифы
		// \p{Hangul} - корейский алфавит
		// \p{Hiragana} и \p{Katakana} - японские слоговые азбуки
		// \p{Bopomofo} - китайская фонетическая азбука
		// .,!?;:'"- - знаки препинания
		// punctuation := `.,!?;:'"-`
		if options.KeepNumbers {
			pattern = `[\p{Han}\p{Hangul}\p{Hiragana}\p{Katakana}\p{Bopomofo}]|[^\p{L}\p{N}\s]`
		} else {
			pattern = `[\p{Han}\p{Hangul}\p{Hiragana}\p{Katakana}\p{Bopomofo}]|[^\p{L}\s]`
		}
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		panic("Failed to compile regexp for mode " + string(mode) + ": " + err.Error())
	}
	return re
}
