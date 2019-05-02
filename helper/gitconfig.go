package helper

import (
	"github.com/mittwald/protodep/logger"
	"gopkg.in/src-d/go-git.v4/plumbing/format/config"
	"os"
	"strings"
)

func GitConfig(target string) string {
	target = strings.TrimSuffix(target, "/")
	cfg := config.New()
	home, err := os.UserHomeDir()
	r, err := os.Open(home + "/.gitconfig")
	if err != nil {
		logger.Error("%v", err)
	}
	decoder := config.NewDecoder(r)
	err = decoder.Decode(cfg)
	if err != nil {
		logger.Error("%v", err)
	}
	for _, subsec := range cfg.Section("url").Subsections {
		if strings.TrimSuffix(subsec.Options.Get("insteadOf"), "/") == target {
			return strings.TrimSuffix(subsec.Name, "/")
		}
	}
	return ""
}
