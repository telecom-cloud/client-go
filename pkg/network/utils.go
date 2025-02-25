package network

import "syscall"

func UnlinkUdsFile(network, addr string) error {
	if network == "unix" {
		return syscall.Unlink(addr)
	}
	return nil
}
