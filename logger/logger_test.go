package logger

import "testing"

func TestCensorHttpsPasswordRemovesPasswordFromHTTPSURLWithPassword(t *testing.T) {
	rawURL := "https://user:password@github.com/foo/bar.git"
	censoredURL := CensorHttpsPassword(rawURL)

	expectedCensoredURL := "https://user:REDACTED@github.com/foo/bar.git"

	if censoredURL != expectedCensoredURL {
		t.Error("incorrectly censored URL", "expected", expectedCensoredURL, "actual", censoredURL)
	}
}

func TestCensorHttpsPasswordKeepsHTTPSURLWithoutPassword(t *testing.T) {
	rawURL := "https://user@github.com/foo/bar.git"
	censoredURL := CensorHttpsPassword(rawURL)

	expectedCensoredURL := "https://user@github.com/foo/bar.git"

	if censoredURL != expectedCensoredURL {
		t.Error("incorrectly censored URL", "expected", expectedCensoredURL, "actual", censoredURL)
	}
}

func TestCensorHttpsPasswordKeepsSSHURLs(t *testing.T) {
	rawURL := "ssh://git@github.com/foo/bar.git"
	censoredURL := CensorHttpsPassword(rawURL)

	expectedCensoredURL := "ssh://git@github.com/foo/bar.git"

	if censoredURL != expectedCensoredURL {
		t.Error("incorrectly censored URL", "expected", expectedCensoredURL, "actual", censoredURL)
	}
}