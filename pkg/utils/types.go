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
		ReplaceYo      bool   `yaml:"replace_yo"` // Замена ё на е
	} `yaml:"cleaner"`

	Logger struct {
		LongWordsLog string `yaml:"long_words_log"` // путь к файлу лога
		Enabled      bool   `yaml:"enabled"`        // включить логирование
	} `yaml:"logger"`

	Tokens struct {
		URLs     bool `yaml:"urls"`
		Emails   bool `yaml:"emails"`
		Numbers  bool `yaml:"numbers"`
		Hashtags bool `yaml:"hashtags"`
		Mentions bool `yaml:"mentions"`
	} `yaml:"tokens"`

	Preserve struct {
		Dates     bool `yaml:"preserve_dates"`
		Fractions bool `yaml:"preserve_fractions"`
		Decimals  bool `yaml:"preserve_decimals"`
	} `yaml:"preserve"`
}

type CleanMode string

const (
	ModeModern      CleanMode = "modern"
	ModeOldSlavonic CleanMode = "old_slavonic"
	ModeAll         CleanMode = "all"
)
