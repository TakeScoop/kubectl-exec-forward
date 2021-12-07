package ports

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/phayes/freeport"
)

// Ports store port mapping information.
type Ports struct {
	Local  int
	Remote interface{}
	Map    string
}

// Parse takes a port mapping (8080:80) and returns a parsed port object, expanding the port 0 to the first open port found.
func Parse(portMap string) (*Ports, error) {
	lstr, remote := splitPort(portMap)

	local, err := strconv.Atoi(lstr)
	if err != nil {
		return nil, err
	}

	if local == 0 {
		local, err = freeport.GetFreePort()
		if err != nil {
			return nil, err
		}
	}

	return &Ports{
		Local:  local,
		Remote: remote,
		Map:    fmt.Sprintf("%d:%s", local, remote),
	}, nil
}

// splitPort is a helper to return ports from a port mapping (8080:80).
func splitPort(port string) (local string, remote string) {
	parts := strings.Split(port, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return parts[0], parts[0]
}
