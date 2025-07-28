package lemmatizer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"
)

type Lemmatizer struct {
	mystemPath  string
	mystemFlags []string
	logFile     *os.File
	logEnabled  bool
	logMutex    sync.Mutex
}

func New(mystemPath, flags string, logEnabled bool, logPath string) (*Lemmatizer, error) {
	lem := &Lemmatizer{
		mystemPath:  mystemPath,
		mystemFlags: parseFlags(flags),
		logEnabled:  logEnabled,
	}

	// Создаем лог-файл только если логирование включено
	if logEnabled {
		// Создаем директории, если их нет
		if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %v", err)
		}

		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
		lem.logFile = file
	}
	return lem, nil
}

func (l *Lemmatizer) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}

func (l *Lemmatizer) Lemmatize(text, filename string) (string, error) {
	if text == "" {
		return "", nil
	}

	// Фильтрация токенов
	var validTokens []string
	tokens := strings.Fields(text)
	for _, token := range tokens {
		// tokenRunes := []rune(token)
		tokenLen := utf8.RuneCountInString(token)

		switch {
		case tokenLen > 100: // Опасные токены
			l.logToken(filename, token, "DANGER")
			continue
		case tokenLen > 30: // Длинные токены
			l.logToken(filename, token, "LONG")
		}
		validTokens = append(validTokens, token)
	}

	filteredText := strings.Join(validTokens, " ")
	if filteredText == "" {
		return "", nil
	}

	// Вызов mystem
	cmd := exec.Command(l.mystemPath, append(l.mystemFlags, "-")...)
	cmd.Stdin = strings.NewReader(filteredText)

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("mystem error: %v, stderr: %s", err, stderr.String())
	}

	return processMystemOutput(out.String()), nil
}

func (l *Lemmatizer) logToken(filename, token, level string) {
	if !l.logEnabled || l.logFile == nil {
		return
	}

	l.logMutex.Lock()
	defer l.logMutex.Unlock()

	logLine := fmt.Sprintf("[%s] %s: %s\n", level, filename, token)
	l.logFile.WriteString(logLine)
}

func processMystemOutput(output string) string {
	var result strings.Builder
	re := regexp.MustCompile(`([^{]*)\{([^|}]+)[^}]*}`)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		// Обрабатываем строки вида: слово{лемма|...} или слово{лемма??}
		for {
			match := re.FindStringSubmatch(line)
			if len(match) == 0 {
				break
			}

			// Добавляем текст перед фигурными скобками
			if match[1] != "" {
				result.WriteString(match[1])
				result.WriteString(" ")
			}

			// Добавляем лемму (убираем ?? и ? в конце)
			lemma := strings.TrimRight(match[2], "?")
			result.WriteString(lemma)
			result.WriteString(" ")

			// Продолжаем обработку оставшейся части строки
			line = line[len(match[0]):]
		}

		// Добавляем оставшийся текст без фигурных скобок
		if line != "" {
			result.WriteString(line)
			result.WriteString(" ")
		}
	}

	// Удаляем лишние пробелы и возвращаем результат
	return strings.Join(strings.Fields(result.String()), " ")
}

func parseFlags(flags string) []string {
	if flags == "" {
		return []string{"-l", "-d"} // дефолтные флаги
	}

	var parsed []string
	for _, f := range flags {
		if f == '-' {
			continue
		}
		parsed = append(parsed, "-"+string(f))
	}
	return parsed
}
