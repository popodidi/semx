package runner

import (
	"fmt"
	"strings"

	"github.com/popodidi/semx/internal/request"
)

func BuildPrompt(req request.RunRequest) string {
	var sections []string
	if system := strings.TrimSpace(req.Prompt.System); system != "" {
		sections = append(sections, "System instructions:\n\n"+system)
	}
	sections = append(sections, fmt.Sprintf(
		"Corpus directory:\n\n%s\n\nInspect files in this directory as necessary.",
		req.Corpus.Path,
	))
	sections = append(sections, "Task:\n\n"+strings.TrimSpace(req.Prompt.User))
	if req.Output.Format == "json" {
		sections = append(sections, fmt.Sprintf(
			"Return only JSON matching the schema at:\n\n%s",
			req.Output.Schema,
		))
	} else {
		sections = append(sections, "Return the evaluation as plain text.")
	}
	return strings.Join(sections, "\n\n")
}
