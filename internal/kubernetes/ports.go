package kubernetes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/phayes/freeport"
)

type Ports struct {
	Local  int
	Remote interface{}
	Map    string
}

func ParsePorts(portMap string) (*Ports, error) {
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

func splitPort(port string) (local string, remote string) {
	parts := strings.Split(port, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return parts[0], parts[0]
}
