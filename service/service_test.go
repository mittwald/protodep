package service

import (
	"github.com/mittwald/protodep/dependency"
	"github.com/mittwald/protodep/helper"
	"github.com/mittwald/protodep/logger"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

// func (s *SyncImpl) getAuthProvider
// (rewrittenGitRepo string, repoURL *url.URL, dep *dependency.ProtoDepDependency, bareDepRepo string)
// (*helper.AuthProvider, error) {

/*
type ProtoDepDependency struct {
	Target   string   `toml:"target"`
	Revision string   `toml:"revision"`
	Branch   string   `toml:"branch"`
	Path     string   `toml:"path"`
	Ignores  []string `toml:"ignores"`
}
*/

/*
Target:   bareDepRepo,
			Branch:   repo.Dep.Branch,
			Revision: repo.Hash,
			Path:     repo.Dep.Path,
			Ignores:  repo.Dep.Ignores,
*/

func TestGetAuthProvider(t *testing.T) {

	dep := dependency.ProtoDepDependency{
		Target:   `target="ply.github.come736765"\n`,
		Revision: `revision="ec1a70913e5793a7d0a7b5fbf7e0e4f75409dd41"\n`,
		Branch:   `branch="master"\n`,
		Path:     `path=""\n`,
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

	for i, v := range tests {
		provider, err := s.getAuthProvider(v.rewrittenGitRepo, v.repoURL, v.dep, v.bareDepRepo)
		if err != nil {
			t.Error(err)
			t.Failed()
		}

		t.Log("Test nr.", i, "Provider :", provider, "Error:", err)

		switch p := (provider).(type) {
		case *helper.AuthProviderHTTPS:
			logger.Info("%#v", p)
			assert.IsType(t, v.typ, p, "REEE")
		case *helper.AuthProviderWithSSH:
			logger.Info("%#v", p)
			assert.IsType(t, v.typ, p, "REEE")
		default:
			t.Log(p)
			t.Failed()
		}
	}

}
