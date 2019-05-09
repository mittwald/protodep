package helper

import (
	"github.com/mittwald/protodep/logger"
	"gopkg.in/src-d/go-git.v4/plumbing/format/config"
	"io"
	"os"
	"strings"
)

func GitConfig(target string) (string, error) {
	target = strings.TrimSuffix(target, "/")
	cfg := config.New()

	r := *loadGitFileFromHome()

	decoder := config.NewDecoder(r)
	err := decoder.Decode(cfg)
	if err != nil {
		logger.Error("%v", err)
	}
	for _, subsec := range cfg.Section("url").Subsections {
		if strings.TrimSuffix(subsec.Options.Get("insteadOf"), "/") == target {
			return strings.TrimSuffix(subsec.Name, "/"), nil
		}
	}
	return "", err
}

func loadGitFileFromHome() *io.Reader {
	home, err := os.UserHomeDir()
	r, err := os.Open(home + "/.gitconfig")
	if err != nil {
		logger.Error("%v", err)
	}

	reader := io.Reader(r)
	return &reader

}
