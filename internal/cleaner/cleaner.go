package cleaner

// TextCleaner определяет интерфейс для очистки текста
type TextCleaner interface {
	Clean(text string) string
}
