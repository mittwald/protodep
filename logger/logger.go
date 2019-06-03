package logger

import (
	"fmt"
	"strings"
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

func CensorHttpsPassword(url string) string {
	path := strings.Split(url, "@")

	if len(path) == 1 {
		return url
	}
	cred := strings.Split(path[0], ":")
	cred[2] = "xxxxxx"
	compCred := strings.Join(cred, ":")

	return compCred + "@" + path[1]
}

func InfoWithSpinner(format string, a ...interface{}) *spinnerWrapper {
	s := spinner.New(spinner.CharSets[38], 100*time.Millisecond) // Build our new spinner
	txt := color.GreenString("[INFO] "+format, a...)
	fmt.Print(txt)
	s.Start()

	return &spinnerWrapper{s}
}
