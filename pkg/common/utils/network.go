package utils

import "net"

const (
	UNKNOWN_IP_ADDR = "-"
)

var localIP string

// LocalIP returns host's ip
func LocalIP() string {
	return localIP
}

// getLocalIp enumerates local net interfaces to find local ip, it should only be called in init phase
func getLocalIp() string {
	inters, err := net.Interfaces()
	if err != nil {
		return UNKNOWN_IP_ADDR
	}
	for _, inter := range inters {
		if inter.Flags&net.FlagLoopback != net.FlagLoopback &&
			inter.Flags&net.FlagUp != 0 {
			addrs, err := inter.Addrs()
			if err != nil {
				return UNKNOWN_IP_ADDR
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					return ipnet.IP.String()
				}
			}
		}
	}

	return UNKNOWN_IP_ADDR
}

func init() {
	localIP = getLocalIp()
}

// TLSRecordHeaderLooksLikeHTTP reports whether a TLS record header
// looks like it might've been a misdirected plaintext HTTP request.
func TLSRecordHeaderLooksLikeHTTP(hdr [5]byte) bool {
	switch string(hdr[:]) {
	case "GET /", "HEAD ", "POST ", "PUT /", "OPTIO":
		return true
	}
	return false
}
