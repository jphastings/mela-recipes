package formats

// LossyField identifies an interchange aspect that a format cannot preserve in a
// particular conversion direction. Reason explains the loss for users, and Present
// reports whether a given recipe actually carries data that would be lost (a nil
// Present means the loss always applies when the format is involved).
type LossyField struct {
	Field   string
	Reason  string
	Present func(InterchangeRecipe) bool
}

// Lossiness declares what a format drops in each conversion direction:
//   - OnImport: converting an interchange recipe into this format (IR -> format).
//   - OnExport: converting a recipe in this format back to interchange (format -> IR).
//
// A format omits a direction entirely when it preserves all content there.
type Lossiness struct {
	OnImport []LossyField
	OnExport []LossyField
}

// ImportLosses returns the fields that would actually be lost converting ir into
// this format: the declared import losses whose data ir actually carries.
func (f *Format) ImportLosses(ir InterchangeRecipe) []LossyField {
	return applicableLosses(f.Lossiness.OnImport, ir)
}

// ExportLosses returns the fields that would actually be lost converting a recipe
// in this format (here described by its interchange form ir) back to interchange.
func (f *Format) ExportLosses(ir InterchangeRecipe) []LossyField {
	return applicableLosses(f.Lossiness.OnExport, ir)
}

func applicableLosses(fields []LossyField, ir InterchangeRecipe) []LossyField {
	var out []LossyField
	for _, lf := range fields {
		if lf.Present == nil || lf.Present(ir) {
			out = append(out, lf)
		}
	}
	return out
}
