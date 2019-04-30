package helper

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4/plumbing/format/config"
	"os"
)

func GitConfig(target string) string{
	cfg := config.New()

	home, err := os.UserHomeDir()
	
	r, err := os.Open(home+"/.gitconfig")
	if err != nil{
		fmt.Println(err)
	}
	decoder := config.NewDecoder(r)
	err = decoder.Decode(cfg)
	if err != nil{
		fmt.Println(err)
	}

	for _, subsec := range cfg.Section("url").Subsections{
		if subsec.Options.Get("insteadOf") == target{
			return subsec.Name
		}
	}
	return ""
}
