package mela

import "github.com/jphastings/recipes/internal/formats"

// SectionedSequence is the shared sectioned ingredient/instruction sequence,
// aliased here so existing code can keep constructing mela.SectionedSequence
// values. The type, its parsing, and the interchange converters live in the
// formats package and are shared with other formats (eg. Paprika).
type SectionedSequence = formats.SectionedSequence
