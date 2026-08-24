package config

type Config struct {
	Runner RunnerConfig `yaml:"runner"`
	Corpus CorpusConfig `yaml:"corpus"`
	Prompt PromptConfig `yaml:"prompt"`
	Output OutputConfig `yaml:"output"`
}

type RunnerConfig struct {
	Type    string         `yaml:"type"`
	Command string         `yaml:"command"`
	Args    map[string]any `yaml:"args"`
}

type CorpusConfig struct {
	Path string `yaml:"path"`
}

type PromptConfig struct {
	System string `yaml:"system"`
	User   string `yaml:"user"`
}

type OutputConfig struct {
	Dir    string `yaml:"dir"`
	Format string `yaml:"format"`
	Schema string `yaml:"schema"`
}

func Defaults() Config {
	return Config{Runner: RunnerConfig{Args: make(map[string]any)}}
}
