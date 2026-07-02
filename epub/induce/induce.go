package induce

import (
	"errors"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

const (
	minUnitCount   = 5    // a delimiter must yield at least this many well-formed recipes
	maxUnitBlocks  = 80   // a single recipe shouldn't be larger than this (guards under-segmentation)
	selMinUnitFrac = 0.30 // a class must appear in this fraction of recipes to be a field
	runPerUnit     = 1.5  // a class averaging at least this many blocks/recipe is a repeated "run"
	ingNumFrac     = 0.30 // a run whose plain lines lead with a numeral this often is ingredients
	proseRatio     = 2.5  // a line this much longer than an ingredient is prose (a step/description)
	// AcceptThreshold is the default gate pass-rate at which a book self-certifies.
	AcceptThreshold = 0.98
)

// wellFormed reports whether a segment looks like exactly one recipe, using only
// structure: a sane number of blocks (not a whole chapter) with a few short
// lines. No recipe vocabulary is consulted.
func wellFormed(u Unit) bool {
	nb := len(u.Blocks)
	if nb < 3 || nb > maxUnitBlocks {
		return false
	}
	short := 0
	for _, b := range u.Blocks {
		if l := runeLen(text(b)); l > 0 && l < 100 {
			short++
		}
	}
	return short >= 2
}

func filter(units []Unit, keep func(Unit) bool) []Unit {
	var out []Unit
	for _, u := range units {
		if keep(u) {
			out = append(out, u)
		}
	}
	return out
}

// candidates returns repeated heading/div classes that could delimit recipes,
// ordered by frequency then by first appearance so selection is deterministic
// and prefers the top-of-recipe heading on ties (e.g. h2.h2a over h2.h2b).
func candidates(docs []Document) []Sel {
	count := map[Sel]int{}
	first := map[Sel]int{}
	ord := 0
	for _, d := range docs {
		eachElement(d.Root, func(n *html.Node) {
			ord++
			if cls := classOf(n); cls != "" {
				switch n.Data {
				case "h1", "h2", "h3", "h4", "h5", "h6", "div":
					s := Sel{Tag: n.Data, Class: cls}
					count[s]++
					if _, seen := first[s]; !seen {
						first[s] = ord
					}
				}
			}
		})
	}
	type sc struct {
		s Sel
		n int
	}
	var scs []sc
	for s, n := range count {
		if n >= minUnitCount {
			scs = append(scs, sc{s, n})
		}
	}
	sort.Slice(scs, func(i, j int) bool {
		if scs[i].n != scs[j].n {
			return scs[i].n > scs[j].n
		}
		return first[scs[i].s] < first[scs[j].s]
	})
	var out []Sel
	for i, x := range scs {
		if i >= 40 {
			break
		}
		out = append(out, x.s)
	}
	return out
}

func modeFor(s Sel) UnitMode {
	if s.Tag == "div" {
		return ModeContainer
	}
	return ModeHeading
}

// --- labelling ---------------------------------------------------------------

// Labeler assigns recipe roles (title, ingredients, steps, …) to a book's
// repeated element classes. The default StructuralLabeler is language-agnostic;
// a model-backed labeller can implement this interface for books too irregular
// for structure alone, and its output flows through the same verification gate.
type Labeler interface {
	Label(units []Unit, unit UnitSpec) map[Role]FieldSpec
}

var defaultLabeler Labeler = StructuralLabeler{}

// StructuralLabeler decides roles purely from layout statistics — line length,
// position within the recipe, how many siblings share a class, leading numerals,
// and emphasis — never from recipe words.
type StructuralLabeler struct{}

type classStat struct {
	count       int
	lens        []int
	sumPos      float64
	numericAll  int // lines leading with a numeral (emphasised or not)
	itemCount   int // non-emphasised lines (candidate items)
	itemNumeric int // …of which lead with a numeral
	units       map[*Unit]bool
}

type classInfo struct {
	sel         Sel
	perUnit     float64 // mean blocks of this class per recipe
	medLen      float64 // median line length
	meanPos     float64 // mean normalised position in the recipe (0=start, 1=end)
	itemNumFrac float64 // of its plain (non-emphasised) lines, the fraction leading with a numeral
	numLeadAll  float64 // of all its lines, the fraction leading with a numeral
	support     float64 // fraction of recipes containing the class
}

func classInfos(units []Unit) []classInfo {
	stats := map[Sel]*classStat{}
	for i := range units {
		u := &units[i]
		nb := len(u.Blocks)
		for j, b := range u.Blocks {
			t := text(b)
			s := Sel{Tag: b.Data, Class: classOf(b)}
			cs := stats[s]
			if cs == nil {
				cs = &classStat{units: map[*Unit]bool{}}
				stats[s] = cs
			}
			cs.count++
			cs.lens = append(cs.lens, runeLen(t))
			if nb > 1 {
				cs.sumPos += float64(j) / float64(nb-1)
			}
			num := startsNumeric(t)
			if num {
				cs.numericAll++
			}
			if !emphasised(b) { // emphasised lines are headings/yield, not items
				cs.itemCount++
				if num {
					cs.itemNumeric++
				}
			}
			cs.units[u] = true
		}
	}
	nU := float64(len(units))
	var out []classInfo
	for s, cs := range stats {
		support := float64(len(cs.units)) / nU
		if support < selMinUnitFrac {
			continue
		}
		numFrac := 0.0
		if cs.itemCount > 0 {
			numFrac = float64(cs.itemNumeric) / float64(cs.itemCount)
		}
		out = append(out, classInfo{
			sel:         s,
			perUnit:     float64(cs.count) / nU,
			medLen:      median(cs.lens),
			meanPos:     cs.sumPos / float64(cs.count),
			itemNumFrac: numFrac,
			numLeadAll:  float64(cs.numericAll) / float64(cs.count),
			support:     support,
		})
	}
	return out
}

func (StructuralLabeler) Label(units []Unit, unit UnitSpec) map[Role]FieldSpec {
	infos := classInfos(units)
	fields := map[Role]FieldSpec{}
	if len(infos) == 0 {
		return fields
	}
	overallMed := medianFloat(medLens(infos))

	// 1. Title is the recipe delimiter (heading mode) or the earliest short,
	//    once-per-recipe line (container mode).
	titleSel := unit.Sel
	if unit.Mode == ModeContainer {
		titleSel = earliestShortSingleton(infos, overallMed)
	}
	if titleSel != (Sel{}) {
		fields[RoleTitle] = FieldSpec{Sels: []Sel{titleSel}}
	}

	// 2. Ingredients are repeated runs whose plain lines mostly lead with a
	//    numeral (a quantity). This wordlessly separates ingredient lists from
	//    image captions and nutrition blocks that also repeat.
	var ing []classInfo
	for _, c := range infos {
		if c.sel == titleSel {
			continue
		}
		if c.perUnit >= runPerUnit && c.itemNumFrac >= ingNumFrac {
			ing = append(ing, c)
		}
	}
	if len(ing) == 0 { // fallback (e.g. all-emphasised ingredients): the most
		// numeral-led short run — in a two-column layout that's the quantity
		// column, which the bare-quantity merge then joins to the name column.
		var best *classInfo
		for i := range infos {
			c := infos[i]
			if c.sel == titleSel || c.perUnit < runPerUnit || c.medLen > overallMed {
				continue
			}
			if best == nil || c.numLeadAll > best.numLeadAll ||
				(c.numLeadAll == best.numLeadAll && c.perUnit > best.perUnit) {
				best = &c
			}
		}
		if best != nil {
			ing = []classInfo{*best}
		}
	}
	inIng := map[Sel]bool{}
	for _, c := range ing {
		inIng[c.sel] = true
	}
	ingMed := refLen(ing, overallMed)
	ingPos := meanPos(ing)
	proseCut := proseRatio * ingMed

	// A distinctly-styled first (or last) ingredient often has its own class that
	// occurs once per recipe, so it isn't a "run" — fold in such a class when it
	// leads with a numeral (a quantity), is ingredient-length (not prose), and
	// sits with the ingredient run (eg. a book's ingredient1 before ingredient).
	// ingMed/ingPos above stay run-based so this outlier doesn't skew the
	// description-vs-method split below.
	for _, c := range infos {
		if inIng[c.sel] || c.sel == titleSel {
			continue
		}
		if c.perUnit < runPerUnit && c.numLeadAll >= ingNumFrac &&
			c.medLen <= proseCut && absFloat(c.meanPos-ingPos) <= 0.3 {
			ing = append(ing, c)
			inIng[c.sel] = true
		}
	}

	// 3. Of the rest: prose (much longer than an ingredient) before the
	//    ingredients is the description, after them is the method; a short
	//    once-per-recipe line before them is a subtitle.
	var steps []classInfo
	var desc, subtitle *classInfo
	for _, c := range infos {
		if c.sel == titleSel || inIng[c.sel] {
			continue
		}
		switch {
		case c.medLen >= proseCut && c.meanPos <= ingPos:
			desc = earlier(desc, c)
		case c.medLen >= proseCut:
			steps = append(steps, c)
		case c.perUnit <= 1.4 && c.meanPos <= ingPos:
			subtitle = earlier(subtitle, c)
		}
	}

	if s := sels(ing); len(s) > 0 {
		fields[RoleIngredients] = FieldSpec{Sels: s, Multiple: true}
	}
	if s := sels(steps); len(s) > 0 {
		fields[RoleSteps] = FieldSpec{Sels: s, Multiple: true}
	}
	if desc != nil {
		fields[RoleDescription] = FieldSpec{Sels: []Sel{desc.sel}}
	}
	if subtitle != nil {
		fields[RoleSubtitle] = FieldSpec{Sels: []Sel{subtitle.sel}}
	}
	return fields
}

func earliestShortSingleton(infos []classInfo, overallMed float64) Sel {
	var best classInfo
	found := false
	for _, c := range infos {
		if c.perUnit <= 1.4 && c.medLen <= overallMed && (!found || c.meanPos < best.meanPos) {
			best, found = c, true
		}
	}
	if found {
		return best.sel
	}
	return Sel{}
}

func firstMatch(u Unit, f FieldSpec) *html.Node {
	for _, b := range u.Blocks {
		if f.Matches(b) {
			return b
		}
	}
	return nil
}

func hasMatch(u Unit, f FieldSpec) bool { return firstMatch(u, f) != nil }

// fieldCoverage records, per role, how often its selector resolves — diagnostic
// detail shown to the user. The accept/flag decision uses the gate pass-rate.
func fieldCoverage(units []Unit, fields map[Role]FieldSpec) map[Role]float64 {
	per := map[Role]float64{}
	n := float64(len(units))
	for _, r := range []Role{RoleTitle, RoleSubtitle, RoleDescription, RoleIngredients, RoleSteps} {
		f, ok := fields[r]
		if !ok || n == 0 {
			per[r] = 0
			continue
		}
		hit := 0
		for _, u := range units {
			if hasMatch(u, f) {
				hit++
			}
		}
		per[r] = float64(hit) / n
	}
	return per
}

// gatePassRate is the fraction of recipes that extract cleanly (verbatim +
// structural invariants). It is the real confidence signal: a wrong labelling
// makes fields "resolve" but fails the gate.
func gatePassRate(unit UnitSpec, fields map[Role]FieldSpec, units []Unit) float64 {
	if len(units) == 0 {
		return 0
	}
	p := &Profile{Unit: unit, Fields: fields}
	ok := 0
	for _, u := range units {
		if p.buildRecipe(u).OK() {
			ok++
		}
	}
	return float64(ok) / float64(len(units))
}

var markerCategory = map[string]string{
	"v": "Vegetarian", "ve": "Vegan", "vg": "Vegan",
	"gf": "Gluten Free", "df": "Dairy Free",
}

// detectMarkers finds inline conventions (e.g. a "v" superscript) in titles.
// Markers are intrinsically a per-book lexical convention, so a small token map
// is appropriate here even though role detection is structural.
func detectMarkers(units []Unit, fields map[Role]FieldSpec) []MarkerRule {
	type mk struct {
		sel Sel
		tok string
	}
	tally := map[mk]int{}
	var titleFields []FieldSpec
	if f, ok := fields[RoleTitle]; ok {
		titleFields = append(titleFields, f)
	}
	if f, ok := fields[RoleSubtitle]; ok {
		titleFields = append(titleFields, f)
	}
	for _, u := range units {
		for _, f := range titleFields {
			node := firstMatch(u, f)
			if node == nil {
				continue
			}
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if c.Type != html.ElementNode || classOf(c) == "" {
					continue
				}
				tok := strings.ToLower(text(c))
				if _, ok := markerCategory[tok]; ok {
					tally[mk{Sel{Tag: c.Data, Class: classOf(c)}, tok}]++
				}
			}
		}
	}
	threshold := int(0.2*float64(len(units)) + 0.5)
	var out []MarkerRule
	for m, n := range tally {
		if n >= threshold && n >= 2 {
			out = append(out, MarkerRule{
				Sel: m.sel, Equals: m.tok,
				Category: markerCategory[m.tok], StripFromTitle: true,
			})
		}
	}
	return out
}

type candResult struct {
	unit   UnitSpec
	fields map[Role]FieldSpec
	conf   Confidence
	score  int
	by     string
}

// evaluate labels a candidate delimiter with one labeller and scores it by how
// many of its recipes extract cleanly.
func evaluate(docs []Document, unit UnitSpec, by string, lab Labeler) *candResult {
	wf := filter(segment(docs, unit), wellFormed)
	if len(wf) < minUnitCount {
		return nil
	}
	fields := lab.Label(wf, unit)
	pass := gatePassRate(unit, fields, wf)
	return &candResult{
		unit:   unit,
		fields: fields,
		conf:   Confidence{PerField: fieldCoverage(wf, fields), Overall: pass, NRecipes: len(wf)},
		score:  int(pass * float64(len(wf))),
		by:     by,
	}
}

func betterThan(a, b *candResult) bool {
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	return a.score > b.score || (a.score == b.score && a.conf.NRecipes > b.conf.NRecipes)
}

// pickBest tries every candidate delimiter with the given labeller and returns
// the best-scoring result.
func pickBest(docs []Document, by string, lab Labeler) *candResult {
	var best *candResult
	for _, cand := range candidates(docs) {
		unit := UnitSpec{Mode: modeFor(cand), Sel: cand}
		if r := evaluate(docs, unit, by, lab); betterThan(r, best) {
			best = r
		}
	}
	return best
}

// Induce discovers a book's recipe structure with no per-book configuration,
// using only the structural labeller.
func Induce(docs []Document, book BookIdent) (*Profile, error) {
	return InduceWith(docs, book, nil)
}

// InduceWith is like Induce but, when the structural labelling doesn't reach the
// accept threshold (role-blind classes, prose, non-English layouts), refines it
// with a model-backed labeller. The model is consulted only for books structure
// can't certify, so self-certifying books pay nothing and can't regress.
func InduceWith(docs []Document, book BookIdent, refiner Labeler) (*Profile, error) {
	best := pickBest(docs, "structural", defaultLabeler)
	if refiner != nil && (best == nil || best.conf.Overall < AcceptThreshold) {
		if m := pickBest(docs, "model", refiner); betterThan(m, best) {
			best = m
		}
	}
	if best == nil {
		return nil, errors.New("no repeating recipe unit discovered")
	}
	markers := detectMarkers(filter(segment(docs, best.unit), wellFormed), best.fields)
	return &Profile{
		SchemaVer:  1,
		Book:       book,
		Unit:       best.unit,
		Fields:     best.fields,
		Markers:    markers,
		Confidence: best.conf,
		InducedBy:  best.by,
	}, nil
}

// --- small numeric helpers ---------------------------------------------------

func median(xs []int) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]int(nil), xs...)
	sort.Ints(s)
	n := len(s)
	if n%2 == 1 {
		return float64(s[n/2])
	}
	return float64(s[n/2-1]+s[n/2]) / 2
}

func medianFloat(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

func medLens(infos []classInfo) []float64 {
	out := make([]float64, len(infos))
	for i, c := range infos {
		out[i] = c.medLen
	}
	return out
}

func refLen(ing []classInfo, fallback float64) float64 {
	if len(ing) == 0 {
		return fallback
	}
	return medianFloat(medLens(ing))
}

func meanPos(cs []classInfo) float64 {
	if len(cs) == 0 {
		return 0.5
	}
	sum := 0.0
	for _, c := range cs {
		sum += c.meanPos
	}
	return sum / float64(len(cs))
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func sels(cs []classInfo) []Sel {
	out := make([]Sel, 0, len(cs))
	for _, c := range cs {
		out = append(out, c.sel)
	}
	return out
}

func earlier(cur *classInfo, c classInfo) *classInfo {
	if cur == nil || c.meanPos < cur.meanPos {
		cc := c
		return &cc
	}
	return cur
}
