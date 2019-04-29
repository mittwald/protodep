package helper

import (
	"fmt"

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
	ep, err := transport.NewEndpoint("ssh://" + repoName + ".git" + ":" + p.port)
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
	var url string
	if len(p.username) > 0 && len(p.password) > 0 {
		url = fmt.Sprintf("https://%s:%s@%s.git", p.username, p.password, repoName)
	} else {
		url = fmt.Sprintf("https://%s.git", repoName)
	}
	return url
}

func (p *AuthProviderHTTPS) AuthMethod() transport.AuthMethod {
	// nil is ok.
	return nil
}
