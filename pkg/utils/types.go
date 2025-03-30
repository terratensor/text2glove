package utils

type Config struct {
	InputDir     string `yaml:"input"`
	OutputFile   string `yaml:"output"`
	WorkersCount int    `yaml:"workers"`
	BufferSize   int    `yaml:"buffer_size"`
	ReportEvery  int    `yaml:"report_every"`

	Cleaner struct {
		Mode           string `yaml:"mode"`
		Normalize      bool   `yaml:"normalize"`
		PreserveSpaces bool   `yaml:"preserve_spaces"`
	} `yaml:"cleaner"`

	Logger struct {
		LongWordsLog string `yaml:"long_words_log"` // путь к файлу лога
		Enabled      bool   `yaml:"enabled"`        // включить логирование
	} `yaml:"logger"`
}

type CleanMode string

const (
	ModeModern      CleanMode = "modern"
	ModeOldSlavonic CleanMode = "old_slavonic"
	ModeAll         CleanMode = "all"
)
