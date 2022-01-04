package command

// Outputs is a collection of Output structs.
type Outputs map[string]Output

// Stdout returns the output Stdout by the Output ID.
func (o Outputs) Stdout() map[string]string {
	out := map[string]string{}

	for id, output := range o {
		out[id] = output.Stdout
	}

	return out
}
