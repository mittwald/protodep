package service

import (
	"github.com/mittwald/protodep/dependency"
	"github.com/mittwald/protodep/helper"
	"github.com/stretchr/testify/assert"
	"io"
	"net/url"
	"strings"
	"testing"
)

func gitConfigReader() *io.Reader {

	str := `
[color]
	status = 1
	log = 1
	branch = 1
	diff = auto

[url "https://user:password@github.com/"]
    insteadOf = https://github.com/`

	r := strings.NewReader(str)

	ior := io.Reader(r)

	return &ior

}

func TestGitConfig(t *testing.T) {

	tests := []struct {
		target            string
		expectEmptyString bool
	}{
		{
			target:            "https://github.com",
			expectEmptyString: false,
		},
		{
			target:            "https://gitlab.com",
			expectEmptyString: true,
		},
	}

	for _, v := range tests {
		rewrittenGitRepo, err := helper.GitConfig(v.target, gitConfigReader())
		if err != nil {
			t.Failed()
		}

		assert.IsType(t, "", rewrittenGitRepo, "Idk, just fk tests i guess")

		if !v.expectEmptyString {
			assert.True(t, len(rewrittenGitRepo) > 0, "string is empty")

		}
	}

}

func TestGetAuthProvider(t *testing.T) {

	dep := dependency.ProtoDepDependency{
		Target:   "",
		Revision: `revision="ec1a70913e5793a7d0a7b5fbf7e0e4f75409dd41"`,
		Branch:   `branch="master"`,
		Path:     `path=""`,
		Ignores:  nil,
	}

	tests := []struct {
		rewrittenGitRepo string
		repoURL          *url.URL
		dep              *dependency.ProtoDepDependency
		bareDepRepo      string
		typ              helper.AuthProvider
	}{
		{
			"",
			&url.URL{
				Scheme:     "https",
				User:       nil,
				Host:       "github.com",
				Path:       "/protocolbuffers/protobuf",
				ForceQuery: false,
			},
			&dep,
			"ply.github.come736765",
			&helper.AuthProviderHTTPS{},
		},
		{
			"ssh://github.com/fffunke/protodep",
			&url.URL{
				Scheme:     "ssh",
				User:       nil,
				Host:       "github.com",
				Path:       "",
				ForceQuery: false,
			},
			&dep,
			"ply.github.come736765",
			&helper.AuthProviderWithSSH{},
		},
		{
			"somerepo3",
			&url.URL{
				Scheme:     "https",
				User:       nil,
				Host:       "github.com",
				Path:       "/protocolbuffers/protobuf",
				ForceQuery: false,
			},
			&dep,
			"ply.github.come736765",
			&helper.AuthProviderHTTPS{},
		},
	}

	ssh := helper.NewSSHAuthProvider("", "", "22")

	https := helper.NewHTTPSAuthProvider("", "")

	s := SyncImpl{
		ssh,
		https,
		"",
		"",
		"",
	}

	for _, v := range tests {
		provider, err := s.getAuthProvider(v.rewrittenGitRepo, v.repoURL, v.dep, v.bareDepRepo)
		if err != nil {
			t.Error(err)
			t.Failed()
		}

		switch p := (provider).(type) {
		case *helper.AuthProviderHTTPS:
			assert.IsType(t, v.typ, p, "didnt get the correct provider")
		case *helper.AuthProviderWithSSH:
			assert.IsType(t, v.typ, p, "didnt get the correct provider")
		default:
			t.Log(p)
			t.Failed()
		}
	}

}
