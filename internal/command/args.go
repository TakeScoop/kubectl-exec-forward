package command

// Args store command specific arguments to be passed to the hook commands.
type Args map[string]string

// Merge merges the provided overrides into the existing args, mutating the existing args.
func (a Args) Merge(overrides map[string]string) {
	for k, v := range overrides {
		a[k] = v
	}
}
