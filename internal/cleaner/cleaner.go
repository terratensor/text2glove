package cleaner

import (
	"regexp"
	"strings"
)

type TextCleaner struct {
	re *regexp.Regexp
}

func New() *TextCleaner {
	return &TextCleaner{
		re: createCleanupRegexp(),
	}
}

func (c *TextCleaner) Clean(text string) string {
	text = strings.ToLower(text)
	text = c.re.ReplaceAllString(text, "")
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

func createCleanupRegexp() *regexp.Regexp {
	pattern := `[^а-яА-ЯёЁ0-9\s]`
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err) // should not happen with hardcoded pattern
	}
	return re
}
