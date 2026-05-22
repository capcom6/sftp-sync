package ftp_test

import (
	"net/url"
	"testing"
)

func TestURLPathExtraction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawURL   string
		wantPath string
	}{
		{
			name:     "nested path",
			rawURL:   "ftp://user:pass@host:21/uploads/project/files",
			wantPath: "/uploads/project/files",
		},
		{name: "single directory", rawURL: "ftp://user:pass@host:21/data", wantPath: "/data"},
		{name: "root path", rawURL: "ftp://user:pass@host:21/", wantPath: "/"},
		{name: "no trailing slash", rawURL: "ftp://user@host:21/path", wantPath: "/path"},
		{name: "with port", rawURL: "ftp://user:pass@host:2121/path/to/dir", wantPath: "/path/to/dir"},
		{name: "empty path", rawURL: "ftp://user:pass@host:21", wantPath: ""},
		{name: "path with special chars", rawURL: "ftp://user:pass@host:21/my%20folder", wantPath: "/my folder"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u, err := url.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("url.Parse(%q) error = %v", tt.rawURL, err)
			}
			if u.Path != tt.wantPath {
				t.Errorf("path = %q, want %q", u.Path, tt.wantPath)
			}
		})
	}
}
