package lemmatizer

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Lemmatizer struct {
	mystemPath  string
	mystemFlags []string
}

func New(mystemPath string, flags string) (*Lemmatizer, error) {
	if _, err := exec.LookPath(mystemPath); err != nil {
		return nil, fmt.Errorf("mystem not found at %s: %v", mystemPath, err)
	}
	return &Lemmatizer{
		mystemPath:  mystemPath,
		mystemFlags: parseFlags(flags),
	}, nil
}

func (l *Lemmatizer) Lemmatize(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	cmd := exec.Command(l.mystemPath, append(l.mystemFlags, "-")...)

	cmd.Stdin = strings.NewReader(text)
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("mystem error: %v, stderr: %s", err, stderr.String())
	}

	return processMystemOutput(out.String()), nil
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
