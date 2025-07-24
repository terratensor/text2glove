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

	Lemmatization struct {
		Enable      bool   `yaml:"enable"`
		MystemPath  string `yaml:"mystem_path"`
		MystemFlags string `yaml:"mystem_flags"`
	} `yaml:"lemmatization"`
}
