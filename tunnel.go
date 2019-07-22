package main

const bufSize = 256

// TunnelData is the data transmitted through reverse tunnel.
type TunnelData struct {
	id   int
	size int
	data []byte
}

// ParseTData parses tunnel data.
func ParseTData(td TunnelData) {

}
