package k8s

import (
	"errors"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

const (
	dynamicPortOffset uint32 = 49152
	dynamicPortMax    uint32 = 65535
)

func createProxyPorts(infs []*capsule.Interface) ([]uint32, error) {
	existingPort := map[uint32]struct{}{}
	for _, inf := range infs {
		existingPort[inf.GetPort()] = struct{}{}
	}

	proxyPorts := make([]uint32, len(infs))

	o := dynamicPortOffset
	for i := range infs {
		var pp uint32
		for {
			pp = o
			if pp > dynamicPortMax {
				return nil, errors.New("could not find an avaiable port in the dynamic range")
			}
			if _, ok := existingPort[pp]; !ok {
				o++
				break
			}
			o++
		}
		proxyPorts[i] = pp
	}

	return proxyPorts, nil
}
