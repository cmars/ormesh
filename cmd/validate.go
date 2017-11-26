package cmd

import (
	"encoding/base64"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	validateClientNameRE = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]+$")
	validateRemoteNameRE = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]+$")
)

func IsValidClientName(name string) bool {
	return validateClientNameRE.MatchString(name)
}

func IsValidRemoteName(name string) bool {
	return validateRemoteNameRE.MatchString(name)
}

func NormalizeAddrPort(addr string) (string, error) {
	if host, port, err := net.SplitHostPort(addr); err == nil {
		if host == "" {
			host = "localhost"
		}
		portNum, err := strconv.Atoi(port)
		if err != nil {
			return "", errors.Wrapf(err, "invalid port %q", port)
		}
		if portNum < 1 || portNum > 65535 {
			return "", errors.Errorf("invalid port %q", port)
		}
		ipAddr, err := net.ResolveIPAddr("ip4", host)
		if err != nil {
			return "", errors.Wrapf(err, "failed to resolve %q", addr)
		}
		return fmt.Sprintf("%s:%d", ipAddr.String(), portNum), nil
	}
	port, err := strconv.Atoi(addr)
	if err != nil {
		return "", errors.Wrapf(err, "invalid address %q", addr)
	}
	return fmt.Sprintf("127.0.0.1:%d", port), nil
}
