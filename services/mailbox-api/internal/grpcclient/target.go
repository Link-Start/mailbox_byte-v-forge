package grpcclient

import (
	"net"
	"strings"
)

func SelfTarget(listenAddr string) string {
	addr := strings.TrimSpace(listenAddr)
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil || port == "" {
		return addr
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}

func TargetPassthrough(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" || strings.Contains(addr, "://") || strings.HasPrefix(addr, "passthrough:") {
		return addr
	}
	return "passthrough:///" + addr
}
