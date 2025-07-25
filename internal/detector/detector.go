package detector

import (
	"log"
	"unicode"
	"unicode/utf8"
)

func IsCorrupted(text string) bool {
	// Если текст невалидный UTF-8
	if !utf8.ValidString(text) {
		log.Printf("Invalid UTF-8 string: %s", text)
		return true
	}

	// Если содержит нулевые байты
	if containsNull(text) {
		log.Printf("String contains null bytes: %s", text)
		return true
	}

	// Если много управляющих символов
	if countControlChars(text) > len(text)/10 { // Более 10%
		log.Printf("String contains too many control characters: %s", text)
		return true
	}

	// Если много символов замены Unicode (�)
	if countReplacementChars(text) > len(text)/20 { // Более 5%
		log.Printf("String contains too many replacement characters: %s", text)
		return true
	}

	return false
}

func containsNull(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return true
		}
	}
	return false
}

func countControlChars(s string) int {
	count := 0
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			count++
		}
	}
	return count
}

func countReplacementChars(s string) int {
	count := 0
	for _, r := range s {
		if r == '\uFFFD' { // Символ замены Unicode �
			count++
		}
	}
	return count
}
