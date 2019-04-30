package helper

import (
	"bufio"
	"fmt"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"net/url"
	"os"
	"strings"

	"github.com/mittwald/protodep/logger"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type AuthProvider interface {
	GetRepositoryURL(repoName string) string
	AuthMethod() transport.AuthMethod
}

type AuthProviderWithSSH struct {
	pemFile  string
	password string
	port     string
}

type AuthProviderHTTPS struct {
	username string
	password string
}

func NewHTTPSAuthProvider(username, password string) AuthProvider {
	logger.Info("use HTTP/HTTPS protocol")
	return &AuthProviderHTTPS{
		username: username,
		password: password,
	}
}

func NewSSHAuthProvider(pemFile, password, port string) AuthProvider {
	logger.Info("use SSH protocol")
	return &AuthProviderWithSSH{
		pemFile:  pemFile,
		password: password,
		port:     port,
	}
}

func (p *AuthProviderWithSSH) GetRepositoryURL(repoName string) string {
	hostname := strings.Split(repoName, "/")[0]
	repoNameWithPort := strings.Replace(repoName, hostname, hostname+":"+p.port, 1)
	ep, err := transport.NewEndpoint(repoNameWithPort + ".git")
	if err != nil {
		panic(err)
	}
	return ep.String()
}

func (p *AuthProviderWithSSH) AuthMethod() transport.AuthMethod {
	am, err := ssh.NewPublicKeysFromFile("git", p.pemFile, p.password)
	if err != nil {
		panic(err)
	}
	return am
}

func (p *AuthProviderHTTPS) GetRepositoryURL(repoName string) string {
	var defaultRepo = fmt.Sprintf("https://%s.git", repoName)
	repoHostname := strings.Split(repoName, "/")[0]

	if len(p.username) > 0 && len(p.password) > 0 {
		return defaultRepo
	}

	homeDir, _ := os.UserHomeDir()
	gitConfig := homeDir + "/.git-credentials"
	if _, err := os.Stat(gitConfig); err != nil {
		logger.Info("... no git-credentials for repo found")
		return defaultRepo
	}

	file, err := os.Open(gitConfig)
	if err != nil {
		logger.Error("%v", err)
		return defaultRepo
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fullCredLine := scanner.Text()

		splitEntry := strings.Split(fullCredLine, "@")
		splitEntryLen := len(splitEntry)
		if splitEntryLen > 2 {
			continue
		}
		gitConfigHostname := splitEntry[splitEntryLen-1]

		if gitConfigHostname != repoHostname {
			continue
		}

		u, err := url.Parse(fullCredLine)
		if err != nil {
			logger.Error("%v", err)
			continue
		}

		if len(u.User.String()) <= 0 {
			continue
		}

		hostnameWithCreds := fmt.Sprintf("%s@%s", u.User.String(), repoHostname)
		repoUrlWithCreds := strings.Replace(defaultRepo, repoHostname, hostnameWithCreds, 1)
		return repoUrlWithCreds
	}

	if err := scanner.Err(); err != nil {
		logger.Error("%v", err)
		return defaultRepo
	}

	return defaultRepo
}

func (p *AuthProviderHTTPS) AuthMethod() transport.AuthMethod {
	if len(p.username) > 0 && len(p.password) > 0 {
		return &http.BasicAuth{
			Username: p.username,
			Password: p.password,
		}
	}
	return nil
}
