package command

// Outputs is a collection of command outputs keyed by ID.
type Outputs map[string]string

// Append copies the existing output, appending the new output, and returns the new, extended outputs.
func (o Outputs) Append(id string, output string) Outputs {
	outputs := Outputs{}

	for k, v := range o {
		outputs[k] = v
	}

	outputs[id] = output

	return outputs
}
