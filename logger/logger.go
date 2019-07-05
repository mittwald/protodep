package logger

import (
	"fmt"
	"net/url"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

func Info(format string, a ...interface{}) {
	color.Green("[INFO] "+format, a...)
}

func Error(format string, a ...interface{}) {
	color.Red("[ERROR] "+format, a...)
}

type spinnerWrapper struct {
	*spinner.Spinner
}

func (s *spinnerWrapper) Finish() {
	s.Stop()
	fmt.Print("\n")
}

func CensorHttpsPassword(repoURL string) string {
	parsed, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Sprintf("<invalid url: %s>", err.Error())
	}

	if parsed.User != nil {
		_, pwdSet := parsed.User.Password()
		if pwdSet {
			parsed.User = url.UserPassword(parsed.User.Username(), "REDACTED")
		}
	}

	return parsed.String()
}

func InfoWithSpinner(format string, a ...interface{}) *spinnerWrapper {
	s := spinner.New(spinner.CharSets[38], 100*time.Millisecond) // Build our new spinner
	txt := color.GreenString("[INFO] "+format, a...)
	fmt.Print(txt)
	s.Start()

	return &spinnerWrapper{s}
}
