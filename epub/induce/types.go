// Package induce induces a per-book extraction profile from an EPUB cookbook's
// own structure, then applies it deterministically to capture recipes verbatim.
//
// The model (if any) is used only to label sample nodes during induction; its
// output is distilled into selectors and discarded. Everything on the hot path
// is deterministic, so extracted text is a literal slice of the source DOM and
// the result self-validates by consistency across the book.
package induce

import (
	"encoding/json"
	"os"

	"golang.org/x/net/html"
)

// Role names a semantic part of a recipe.
type Role string

const (
	RoleTitle       Role = "title"
	RoleSubtitle    Role = "subtitle"
	RoleDescription Role = "description"
	RoleIngredients Role = "ingredients"
	RoleSteps       Role = "steps"
	RoleYield       Role = "yield"
)

// Sel matches an element by tag and/or class. An empty field is a wildcard, but
// at least one must be set to match anything.
type Sel struct {
	Tag   string `json:"tag,omitempty"`
	Class string `json:"class,omitempty"`
}

func (s Sel) Matches(n *html.Node) bool {
	if n == nil || n.Type != html.ElementNode {
		return false
	}
	if s.Tag == "" && s.Class == "" {
		return false
	}
	if s.Tag != "" && n.Data != s.Tag {
		return false
	}
	// An empty Class means the classless group specifically, not "any class" —
	// otherwise a classless selector (e.g. p) would match every paragraph.
	return classOf(n) == s.Class
}

// XPath renders the selector in the human-readable form used in profiles.
func (s Sel) XPath() string {
	switch {
	case s.Tag != "" && s.Class != "":
		return "//" + s.Tag + "[@class='" + s.Class + "']"
	case s.Tag != "":
		return "//" + s.Tag + "[not(@class)]"
	case s.Class != "":
		return "//*[@class='" + s.Class + "']"
	}
	return ""
}

func (s Sel) String() string { return s.XPath() }

// UnitMode is how a document is segmented into recipe units.
type UnitMode string

const (
	ModeHeading   UnitMode = "heading"   // a heading and following siblings until the next same heading
	ModeContainer UnitMode = "container" // a container element's subtree is one unit
)

type UnitSpec struct {
	Mode UnitMode `json:"mode"`
	Sel  Sel      `json:"selector"`
}

// FieldSpec captures one role. A role may be carried by several classes
// (e.g. ingredients in both p.ing and p.ing1), so it holds a set of selectors.
type FieldSpec struct {
	Sels     []Sel `json:"selectors"`
	Multiple bool  `json:"multiple,omitempty"`
}

func (f FieldSpec) Matches(n *html.Node) bool {
	for _, s := range f.Sels {
		if s.Matches(n) {
			return true
		}
	}
	return false
}

func (f FieldSpec) XPaths() []string {
	out := make([]string, len(f.Sels))
	for i, s := range f.Sels {
		out[i] = s.XPath()
	}
	return out
}

// MarkerRule maps an inline convention (e.g. a "v" superscript) onto a category.
type MarkerRule struct {
	Sel            Sel    `json:"selector"`
	Equals         string `json:"equals"`
	Category       string `json:"category"`
	StripFromTitle bool   `json:"stripFromTitle"`
}

type BookIdent struct {
	ISBN  string `json:"isbn,omitempty"`
	Title string `json:"title,omitempty"`
}

// Confidence is the measured cross-recipe coverage that drives accept/flag.
type Confidence struct {
	PerField map[Role]float64 `json:"perField"`
	Overall  float64          `json:"overall"`
	NRecipes int              `json:"nRecipes"`
}

// SelfCertifies reports whether enough recipes extracted cleanly (passed the
// verbatim + structural gate) for the book to be trusted without human review.
func (c Confidence) SelfCertifies(threshold float64) bool {
	return c.Overall >= threshold
}

// Profile is the declarative, auto-induced extraction spec for one book. It is
// the only per-book artifact — there is no hand-written Go per book.
type Profile struct {
	SchemaVer  int                `json:"schemaVer"`
	Book       BookIdent          `json:"book"`
	Unit       UnitSpec           `json:"unit"`
	Fields     map[Role]FieldSpec `json:"fields"`
	Markers    []MarkerRule       `json:"markers,omitempty"`
	Confidence Confidence         `json:"confidence"`
	InducedBy  string             `json:"inducedBy"`
}

func (p *Profile) Save(path string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func LoadProfile(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
