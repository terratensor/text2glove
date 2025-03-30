# Text2Glove

[![Go Reference](https://pkg.go.dev/badge/github.com/terratensor/text2glove.svg)](https://pkg.go.dev/github.com/terratensor/text2glove)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/terratensor/text2glove/blob/main/LICENSE)

Инструмент для предобработки текстовых файлов перед обучением векторных представлений слов GloVe.

## Основные возможности

- Поддержка **многоязычных текстов** (русский, английский, европейские языки, турецкий)
- Специальная обработка **старославянских и старорусских текстов**
- Параллельная обработка gz-архивов
- Очистка текста с сохранением:
- Букв (включая специфические символы разных языков)
- Цифр
- Пробелов
- Эффективная потоковая запись результатов
- Мониторинг прогресса обработки
- Гибкая конфигурация через YAML-файлы

## Установка

```bash
go install github.com/terratensor/text2glove/cmd/text2glove@latest
```

## Быстрый старт

Обработка современных текстов:
```bash
text2glove --input ./data --output output.txt --workers 8 --cleaner_mode modern
```

Обработка старославянских текстов:
```bash
text2glove --input ./old_texts --output output.txt --cleaner_mode old_slavonic
```

## Конфигурация

Пример файла `config.yaml`:

```yaml
input: "./data" # Директория с файлами .gz
output: "./output.txt" # Выходной файл
workers: 8 # Количество рабочих процессов
buffer_size: 1048576 # Размер буфера записи (1MB)

cleaner:
mode: "old_slavonic" # Режим очистки: modern|old_slavonic|all
normalize: true # Нормализация Unicode
```

Запуск с конфигурационным файлом:
```bash
text2glove --config config.yaml
```

Запуск с параметрами:
```bash
gtext2glove --input ./data --output out.txt
```

Пример ожидаемого вывода:
```bash
    Processing: [=====================>  ] 85.3% | Speed: 142,305.8 KB/s | Lines: 1,284,567

    === Processing completed ===
    Time:    12m45s
    Lines:   14,201,558
    Data:    12.4 GB
    Speed:   152,304.2 KB/s
```

## Режимы обработки текста

1. **modern** - современные языки:
- Русский (включая ё)
- Английский и основные европейские языки
- Турецкий

2. **old_slavonic** - старославянские тексты:
- Все современные символы
- Старославянская кириллица (ѣ, ѵ, ѳ и др.)
- Поддержка Unicode-диапазонов

3. **all** - все Unicode-символы:
- Для специализированных задач

## Сборка из исходников

```bash
git clone https://github.com/terratensor/text2glove
cd text2glove
make build
```

Собранный бинарный файл будет доступен в `bin/text2glove`

## Примеры обработки

Современный русский:
```text
"Привет, мир!" → "привет мир"
```

Старославянский:
```text
"Цѣрь града сего" → "цѣрь града сего"
```

Смешанный текст:
```text
"Сіе есть modern text" → "сіе есть modern text"
```

## Вклад в проект

PR и предложения приветствуются! Пожалуйста:
1. Создайте issue для обсуждения изменений
2. Опишите предлагаемую функциональность
3. Убедитесь, что код проходит все тесты

## Лицензия

MIT License. Подробнее см. в файле [LICENSE](LICENSE).