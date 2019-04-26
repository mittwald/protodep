package helper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRepositoryURLWithSSH(t *testing.T) {
	target := &AuthProviderWithSSH{}
	actual := target.GetRepositoryURL("github.com/mittwald/protodep")

	require.Equal(t, "ssh://github.com/mittwald/protodep.git", actual)
}

func TestGetRepositoryURLHTTPS(t *testing.T) {
	target := &AuthProviderHTTPS{}
	actual := target.GetRepositoryURL("github.com/mittwald/protodep")

	require.Equal(t, "https://github.com/mittwald/protodep.git", actual)
}
