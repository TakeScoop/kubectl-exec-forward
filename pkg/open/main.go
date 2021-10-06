package open

import (
	"os/exec"
)

func Open(address string) error {
	cmd := exec.Command("open", address)

	_, err := cmd.Output()
	if err != nil {
		return err
	}

	return nil
}
