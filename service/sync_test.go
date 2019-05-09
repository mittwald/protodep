package service

import (
	"github.com/mittwald/protodep/helper"
	"github.com/mittwald/protodep/logger"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"testing"
)

var authMethod transport.AuthMethod

var authProvider helper.AuthProvider = helper.AuthProvider{

}



func TestNewSync(t *testing.T) {

	authProviderSSH := authProvider
	authProviderHTTPS := authProvider
	userHomeDir := "/home/ffunke"
	targetDir := "/home/ffunke/git/protodep"
	outputRootDir := "/home/ffunke/git/protodep"

	expected := &SyncImpl{
		authProviderSSH:   authProviderSSH,
		authProviderHTTPS: authProviderHTTPS,
		userHomeDir:       userHomeDir,
		targetDir:         targetDir,
		outputRootDir:     outputRootDir,
	}

	sync := NewSync(authProviderSSH, authProviderHTTPS, userHomeDir, targetDir, outputRootDir)

	if !assert.Equal(t, expected, sync) {
		logger.Info("%#v\n", sync)
		logger.Info("%#v\n", expected)
		t.Errorf("IDFK")
	}

}
