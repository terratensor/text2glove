package customcleaner

import (
	"strings"

	"github.com/terratensor/segment/rule"
	"github.com/terratensor/segment/split"
	"github.com/terratensor/segment/tokenizer"
)

// CustomCleaner реализует TextCleaner интерфейс
type CustomCleaner struct {
}

func NewCustomCleaner() *CustomCleaner {
	return &CustomCleaner{}
}

func (c *CustomCleaner) Clean(text string) string {
	// Своя логика очистки
	text = strings.ToLower(text)

	splitter := split.NewSplitter(3)
	rules := []rule.Rule{
		rule.NewDashRule(),
		rule.NewFloatRule(),
		rule.NewFractionRule(),
		rule.NewUnderscoreRule(),
		rule.NewPunctRule(),
		rule.NewOtherRule(),
		// rule.NewYahooRule(),
	}
	tokenizer := tokenizer.NewTokenizer(splitter, rules)
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
