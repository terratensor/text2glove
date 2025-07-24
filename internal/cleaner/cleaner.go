package cleaner

import (
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

type CleanMode string

const (
	ModeModern         CleanMode = "modern"
	ModeOldSlavonic    CleanMode = "old_slavonic"
	ModeAll            CleanMode = "all"
	ModeUnicodeLetters CleanMode = "unicode_letters"
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
		// Современные языки: русский, английский, основные европейские
		pattern = `[^\p{L}\p{N}\sа-яёa-zà-ÿğüşıöç]`

	case ModeOldSlavonic:
		// Старославянские символы (явное перечисление)
		oldSlavonicChars := "ѣѢѵѴіІѳѲѫѪѭѬѧѦѩѨѯѮѱѰѡѠѿѾҌҍꙋꙊꙗꙖꙙꙘꙜꙛꙝꙞꙟꙠꙡꙢꙣꙤꙥꙦꙧꙨꙩꙪꙫꙬꙭꙮѻѺѹѸѷѶѵѴѳѲѱѰѯѮѭѬѫѪѩѨѧѦѥѤѣѢѣѢѡѠџЏѾѽѼѻѺѹѸ"
		pattern = `[^\p{L}\p{N}\s` + oldSlavonicChars + `]`

	case ModeAll:
		// Все буквы Unicode и цифры
		pattern = `[^\p{L}\p{N}\s]`

	case ModeUnicodeLetters:
		// Все буквы Unicode без цифр
		pattern = `[^\p{L}\s]`
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		panic("Failed to compile regexp for mode " + string(mode) + ": " + err.Error())
	}
	return re
}
