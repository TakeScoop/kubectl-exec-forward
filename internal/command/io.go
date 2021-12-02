package command

import (
	"io"
	"os"
)

type IO struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func NewIO(stdin *io.Reader, stdout *io.Writer, stderr *io.Writer) IO {
	ios := IO{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if stdin != nil {
		ios.Stdin = *stdin
	}

	if stdout != nil {
		ios.Stdout = *stdout
	}

	if stderr != nil {
		ios.Stderr = *stderr
	}

	return ios
}
