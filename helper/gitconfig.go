package helper

import (
	"github.com/mittwald/protodep/logger"
	"gopkg.in/src-d/go-git.v4/plumbing/format/config"
	"io"
	"os"
	"strings"
)

func GitConfig(target string, r *io.Reader) ([]string, error) {
	target = strings.TrimSuffix(target, "/")
	cfg := config.New()
	var rewrites []string

	decoder := config.NewDecoder(*r)
	err := decoder.Decode(cfg)
	if err != nil {
		logger.Error("%v", err)
	}
	for _, subsec := range cfg.Section("url").Subsections {
		if strings.TrimSuffix(subsec.Options.Get("insteadOf"), "/") == target {
			rewrites = append(rewrites, strings.TrimSuffix(subsec.Name, "/"))
		}
	}
	return rewrites, err
}

func LoadGitCredentialsFileFromHome() (*io.Reader, error) {
	homeDir, _ := os.UserHomeDir()
	gitCredentials := homeDir + "/.git-credentials"
	if _, err := os.Stat(gitCredentials); err != nil {
		logger.Info("... no git-credentials found")
		return nil, err
	}

	file, err := os.Open(gitCredentials)
	if err != nil {
		logger.Info("... no git-credentials for repo found")
		return nil, err
	}

	reader := io.Reader(file)

	return &reader, nil
}

func LoadGitConfigFileFromHome() (*io.Reader, error) {
	home, err := os.UserHomeDir()
	r, err := os.Open(home + "/.gitconfig")
	if err != nil {
		logger.Error("%v", err)
		return nil, err
	}

	reader := io.Reader(r)
	return &reader, nil

}
