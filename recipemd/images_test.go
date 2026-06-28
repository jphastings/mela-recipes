package recipemd

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/utils"
)

func pngBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

// TestImageDataURLRoundTrip: an embedded image is written as a base64 data: URL
// and read back offline (no network), preserving the bytes.
func TestImageDataURLRoundTrip(t *testing.T) {
	img := pngBytes(t)
	ir := formats.NewInterchangeRecipe()
	ir.Title = "Photo"
	ir.Images = []utils.B64Image{img}
	ir.Ingredients = ingGroups(ingGroup("", "1 egg"))
	ir.Instructions = sections(sec("", "Cook it."))

	rec, err := Import(ir)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	var buf bytes.Buffer
	if err := rec.Marshal(&buf); err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(buf.String(), "![](data:image/png;base64,") {
		t.Errorf("output missing data: image:\n%s", buf.String())
	}

	got, err := parseRecipe(buf.Bytes(), false)
	if err != nil {
		t.Fatalf("parseRecipe: %v", err)
	}
	if len(got.Images) != 1 {
		t.Fatalf("got %d images, want 1", len(got.Images))
	}
	if !bytes.Equal(got.Images[0], img) {
		t.Error("round-tripped image bytes differ from the original")
	}
}

// TestImageNetworkGating: a remote image is only fetched when network access is
// permitted.
func TestImageNetworkGating(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngBytes(t))
	}))
	defer srv.Close()

	md := fmt.Sprintf("# Soup\n\n![](%s/hero.png)\n\n---\n\n- 1 egg\n\n---\n\n1. Cook it.\n", srv.URL)

	off, err := parseRecipe([]byte(md), false)
	if err != nil {
		t.Fatalf("parseRecipe (off): %v", err)
	}
	if len(off.Images) != 0 {
		t.Errorf("network off: got %d images, want 0", len(off.Images))
	}
	if hits != 0 {
		t.Errorf("network off: server hit %d times, want 0", hits)
	}

	on, err := parseRecipe([]byte(md), true)
	if err != nil {
		t.Fatalf("parseRecipe (on): %v", err)
	}
	if len(on.Images) != 1 {
		t.Fatalf("network on: got %d images, want 1", len(on.Images))
	}
}
