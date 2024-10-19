package crouton

func (r *Recipe) Standardize() error {
	r.Filename = stringToFilename(r.Name)

	// TODO: More

	return nil
}
