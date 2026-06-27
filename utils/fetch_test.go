package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchImage(t *testing.T) {
	const body = "\x89PNG\r\n\x1a\n pretend image bytes"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte(body))
		case "/notimage":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html>not an image</html>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Run("success", func(t *testing.T) {
		img, err := FetchImage(srv.URL + "/ok.png")
		if err != nil {
			t.Fatalf("FetchImage: %v", err)
		}
		if string(img) != body {
			t.Errorf("got %q, want %q", string(img), body)
		}
	})

	t.Run("not found", func(t *testing.T) {
		if _, err := FetchImage(srv.URL + "/missing.png"); err == nil {
			t.Error("expected an error for a 404")
		}
	})

	t.Run("non-image content type", func(t *testing.T) {
		if _, err := FetchImage(srv.URL + "/notimage"); err == nil {
			t.Error("expected an error for a non-image content type")
		}
	})
}
