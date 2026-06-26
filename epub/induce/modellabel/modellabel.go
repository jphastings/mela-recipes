// Package modellabel provides a model-backed induce.Labeler. Where a book's CSS
// classes are role-blind (auto-generated names, tables, prose), the structural
// labeller can't tell an ingredient from a step; this one classifies each
// class's text by meaning using a small multilingual sentence-embedding model,
// so it works across languages. The model never emits recipe text — it only
// labels which class plays which role, which is then distilled into selectors
// and validated by the same verbatim gate.
package modellabel

import (
	"context"
	"math"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/jphastings/recipes/epub/induce"
	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/pipelines"
	"golang.org/x/net/html"
)

// modelRepo is a small multilingual sentence-embedding model that runs on
// Hugot's pure-Go backend (no CGo, no shared libraries).
const modelRepo = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"

const onnxFile = "onnx/model.onnx"

// prototypes anchor each role with a few exemplar lines. Because the embedding
// model is multilingual, an ingredient line in Spanish or French still lands
// near these English exemplars.
var prototypes = map[induce.Role][]string{
	induce.RoleTitle: {
		"Roast chicken with lemon and herbs", "Chocolate fudge cake",
		"Spinach and ricotta tart", "Beef stew with red wine",
	},
	induce.RoleIngredients: {
		"2 cups plain flour", "1 onion, finely chopped", "200 g butter, softened",
		"a pinch of salt", "3 large eggs", "1 tablespoon olive oil",
	},
	induce.RoleSteps: {
		"Heat the oil in a frying pan and cook the onions until soft and golden.",
		"Bake in the oven for 30 minutes until risen and golden.",
		"Whisk the eggs with the sugar until pale and thick.",
		"Bring a large pan of salted water to the boil and cook the pasta.",
	},
	induce.RoleDescription: {
		"This is a classic family recipe passed down through generations.",
		"A light and refreshing dish, perfect for a summer lunch.",
		"My grandmother used to make this every Sunday.",
	},
	// roleOther collects non-recipe furniture — nutrition panels, serving
	// counts, photo credits, cross-references — so those classes are excluded
	// from the recipe fields rather than mistaken for ingredients.
	roleOther: {
		"PER SERVING 267 calories, 12 g protein, 8 g fat, 38 g carbs",
		"Nutrition per portion: 350 kcal",
		"Serves 4", "Makes 12 portions",
		"See page 18 for the method",
		"Photograph by Jane Doe",
	},
}

const roleOther = induce.Role("other")

// Labeler is a model-backed induce.Labeler. Construct one with New and reuse it
// across books; call Close when done.
type Labeler struct {
	session *hugot.Session
	emb     *pipelines.FeatureExtractionPipeline
	roleVec map[induce.Role][]float32
}

var _ induce.Labeler = (*Labeler)(nil)

// New loads the embedding model (downloading it into modelDir on first use) and
// embeds the role prototypes.
func New(ctx context.Context, modelDir string) (*Labeler, error) {
	session, err := hugot.NewGoSession(ctx)
	if err != nil {
		return nil, err
	}
	opts := hugot.NewDownloadOptions()
	opts.OnnxFilePath = onnxFile
	modelPath, err := hugot.DownloadModel(ctx, modelRepo, modelDir, opts)
	if err != nil {
		_ = session.Destroy()
		return nil, err
	}
	emb, err := hugot.NewPipeline(session, hugot.FeatureExtractionConfig{
		ModelPath:    modelPath,
		Name:         "rolelabel",
		OnnxFilename: onnxFile,
	})
	if err != nil {
		_ = session.Destroy()
		return nil, err
	}
	l := &Labeler{session: session, emb: emb, roleVec: map[induce.Role][]float32{}}
	for role, exemplars := range prototypes {
		out, err := emb.RunPipeline(ctx, exemplars)
		if err != nil {
			_ = session.Destroy()
			return nil, err
		}
		l.roleVec[role] = centroid(out.Embeddings)
	}
	return l, nil
}

func (l *Labeler) Close() error { return l.session.Destroy() }

const (
	samplesPerClass = 6
	embedBatch      = 32
)

// Label classifies each repeated element class by the meaning of its lines and
// distils the result into a role->selector map.
func (l *Labeler) Label(units []induce.Unit, unit induce.UnitSpec) map[induce.Role]induce.FieldSpec {
	type acc struct {
		texts []string
		count int
	}
	per := map[induce.Sel]*acc{}
	for _, u := range units {
		for _, b := range u.Blocks {
			t := normText(b)
			if t == "" {
				continue
			}
			sel := induce.Sel{Tag: b.Data, Class: classOf(b)}
			a := per[sel]
			if a == nil {
				a = &acc{}
				per[sel] = a
			}
			a.count++
			if len(a.texts) < samplesPerClass {
				a.texts = append(a.texts, t)
			}
		}
	}
	minOccur := int(0.3*float64(len(units)) + 0.5)
	if minOccur < 2 {
		minOccur = 2
	}

	// Embed all sampled lines in batches, then majority-vote a role per class.
	var texts []string
	var owner []induce.Sel
	for sel, a := range per {
		if a.count < minOccur {
			continue
		}
		for _, t := range a.texts {
			texts = append(texts, t)
			owner = append(owner, sel)
		}
	}
	votes := map[induce.Sel]map[induce.Role]int{}
	ctx := context.Background()
	for i := 0; i < len(texts); i += embedBatch {
		end := i + embedBatch
		if end > len(texts) {
			end = len(texts)
		}
		out, err := l.emb.RunPipeline(ctx, texts[i:end])
		if err != nil {
			continue
		}
		for j, v := range out.Embeddings {
			sel := owner[i+j]
			if votes[sel] == nil {
				votes[sel] = map[induce.Role]int{}
			}
			votes[sel][l.nearest(v)]++
		}
	}

	byRole := map[induce.Role][]induce.Sel{}
	for sel, v := range votes {
		byRole[winner(v)] = append(byRole[winner(v)], sel)
	}

	fields := map[induce.Role]induce.FieldSpec{}
	// The title is the structural delimiter in heading mode; otherwise take the
	// class the model called a title.
	if unit.Mode == induce.ModeHeading {
		fields[induce.RoleTitle] = induce.FieldSpec{Sels: []induce.Sel{unit.Sel}}
	} else if ts := byRole[induce.RoleTitle]; len(ts) > 0 {
		fields[induce.RoleTitle] = induce.FieldSpec{Sels: ts[:1]}
	}
	if s := byRole[induce.RoleIngredients]; len(s) > 0 {
		fields[induce.RoleIngredients] = induce.FieldSpec{Sels: s, Multiple: true}
	}
	if s := byRole[induce.RoleSteps]; len(s) > 0 {
		fields[induce.RoleSteps] = induce.FieldSpec{Sels: s, Multiple: true}
	}
	if s := byRole[induce.RoleDescription]; len(s) > 0 {
		fields[induce.RoleDescription] = induce.FieldSpec{Sels: s[:1]}
	}
	return fields
}

func (l *Labeler) nearest(v []float32) induce.Role {
	best, bestScore := induce.Role(""), float32(-2)
	for role, c := range l.roleVec {
		if s := cosine(v, c); s > bestScore {
			bestScore, best = s, role
		}
	}
	return best
}

func winner(votes map[induce.Role]int) induce.Role {
	best, bestN := induce.Role(""), 0
	for r, n := range votes {
		if n > bestN {
			bestN, best = n, r
		}
	}
	return best
}

func classOf(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "class" {
			return strings.TrimSpace(a.Val)
		}
	}
	return ""
}

func normText(n *html.Node) string {
	return strings.Join(strings.Fields(htmlquery.InnerText(n)), " ")
}

func centroid(vs [][]float32) []float32 {
	if len(vs) == 0 {
		return nil
	}
	c := make([]float32, len(vs[0]))
	for _, v := range vs {
		for i, x := range v {
			c[i] += x
		}
	}
	for i := range c {
		c[i] /= float32(len(vs))
	}
	return c
}

func cosine(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(na) * math.Sqrt(nb)))
}
