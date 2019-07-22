package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

var err error

func main() {
	fmt.Println("*** Reverse Tunnel Slave***")

	tnAddr := flag.String("t", "", "tunnel address")
	ctAddr := flag.String("c", "", "address to be connected")
	flag.Parse()
	if *tnAddr == "" || *ctAddr == "" {
		fmt.Println("Address not specified, please see usage")
		os.Exit(0)
	}

	rts := NewRTSlave(*ctAddr, *tnAddr)
	rts.Start()
}

// RTSlave is the Slave side of reverse tunnel.
type RTSlave struct {
	ctAddr *net.TCPAddr
	tnAddr *net.TCPAddr

	cliConns map[int]net.Conn
	tnConn   net.Conn

	ch chan TunnelData
}

// NewRTSlave returns a RTSlave.
func NewRTSlave(lnAddr, tnAddr string) *RTSlave {
	c, err := net.ResolveTCPAddr("tcp", lnAddr)
	if err != nil {
		fmt.Println("Invilid address")
	}
	t, err := net.ResolveTCPAddr("tcp", tnAddr)
	if err != nil {
		fmt.Println("Invilid address")
	}
	return &RTSlave{
		ctAddr:   c,
		tnAddr:   t,
		cliConns: make(map[int]net.Conn),
		ch:       make(chan TunnelData),
	}
}

// Start .
func (rts *RTSlave) Start() {
	go func() {
		var err error
		rts.tnConn, err = net.DialTCP("tcp", nil, rts.tnAddr)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		fmt.Println("[+] Tunnel established")
		rts.handleTunnelConn()
	}()

	// new tunnel data comes
	for td := range rts.ch {
		// not existed before
		if rts.cliConns[td.id] == nil {
			rts.cliConns[td.id], err = net.DialTCP("tcp", nil, rts.ctAddr)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("[!] Client", td.id, "comes")
		}

		// forward
		n, _ := rts.cliConns[td.id].Write(td.data[:td.size])
		fmt.Println("[*] forward", n, "bytes from client", td.id)
		// wait response and send back
		go rts.handleCliConns(td.id)
	}

}

// read from tunnel and forward
func (rts *RTSlave) handleTunnelConn() {
	for {
		var buf [bufSize]byte
		_, err := rts.tnConn.Read(buf[0:])
		if err != nil {
			return
		}
		td := TunnelData{
			id:   int(buf[0]),
			size: int(buf[1]),
			data: buf[2:],
		}
		rts.ch <- td
	}
}

// read response and send back to tunnel
func (rts *RTSlave) handleCliConns(id int) {
	for {
		// reserve place for id and size
		var buf [bufSize - 2]byte
		n, err := rts.cliConns[id].Read(buf[0:])
		if err != nil {
			return
		}
		tmp := []byte{byte(id), byte(n)}
		tmp = append(tmp, buf[:]...)
		rts.tnConn.Write(tmp)
		fmt.Println("[*] send back", n, "bytes", "to client", id)
	}
}
