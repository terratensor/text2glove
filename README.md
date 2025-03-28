# Text2Glove

[![Go Reference](https://pkg.go.dev/badge/github.com/terratensor/text2glove.svg)](https://pkg.go.dev/github.com/terratensor/text2glove)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/terratensor/text2glove/blob/main/LICENSE)

Tool for preprocessing text files for GloVe word embeddings training.

## Features

- Processes gzipped text files in parallel
- Cleans text (keeps only letters, numbers, spaces)
- Efficient streaming output
- Progress monitoring
- Configurable via YAML

## Installation

```bash
go install github.com/terratensor/text2glove/cmd/text2glove@latest
```

## Quick Start

```bash
text2glove --input ./data --output output.txt --workers 8
```

## Configuration

Create `config.yaml`:

```yaml
input: "./data"
output: "./output.txt"
workers: 8
buffer_size: 1048576
```

Then run:
```bash
text2glove --config config.yaml
```

## Build from Source

```bash
git clone https://github.com/terratensor/text2glove
cd text2glove
make build
```

## Contributing

PRs are welcome! Please open an issue first to discuss proposed changes.