package handlers

import "testing"

func TestBuildSPARedirectMergesParams(t *testing.T) {
	got := buildSPARedirect("http://localhost:5173/path?existing=1", map[string]string{
		"session_id": "abc",
	})

	want := "http://localhost:5173/path?existing=1&session_id=abc"
	if got != want {
		t.Fatalf("redirect = %q, want %q", got, want)
	}
}

func TestBuildSPARedirectRejectsNonHTTP(t *testing.T) {
	if got := buildSPARedirect("file:///tmp/x", map[string]string{"session_id": "abc"}); got != "" {
		t.Fatalf("redirect = %q, want empty", got)
	}
}
