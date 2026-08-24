package request

import "github.com/popodidi/semx/internal/config"

type RunRequest struct {
	Corpus Corpus
	Prompt Prompt
	Output OutputContract
}

type Corpus struct {
	Path string
}

type Prompt struct {
	System string
	User   string
}

type OutputContract struct {
	Format string
	Schema string
}

func FromConfig(cfg config.Config) RunRequest {
	return RunRequest{
		Corpus: Corpus{Path: cfg.Corpus.Path},
		Prompt: Prompt{System: cfg.Prompt.System, User: cfg.Prompt.User},
		Output: OutputContract{Format: cfg.Output.Format, Schema: cfg.Output.Schema},
	}
}
