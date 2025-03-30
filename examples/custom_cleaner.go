package main

import (
	"strings"

	"github.com/terratensor/segment"
)

// CustomCleaner реализует TextCleaner интерфейс
type CustomCleaner struct {
	keepPunctuation bool
}

func NewCustomCleaner(keepPunctuation bool) *CustomCleaner {
	return &CustomCleaner{
		keepPunctuation: keepPunctuation,
	}
}

func (c *CustomCleaner) Clean(text string) string {
	// Своя логика очистки
	text = strings.ToLower(text)

	tokenizer := segment.NewTokenizer()
	tokens := tokenizer.Tokenize(text)

	cleaned_text := make([]string, len(tokens))
	for i, token := range tokens {
		cleaned_text[i] = token.Text
	}
	text = strings.Join(cleaned_text, " ")

	// if !c.keepPunctuation {
	// 	text = removePunctuation(text)
	// }

	return strings.TrimSpace(text)
}

// func removePunctuation(text string) string {
// 	// ... реализация удаления пунктуации
// }
