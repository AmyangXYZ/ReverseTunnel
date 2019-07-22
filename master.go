package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("*** Reverse Tunnel Master***")

	tnAddr := flag.String("t", "", "tunnel address")
	lnAddr := flag.String("l", "", "address to be listened")
	flag.Parse()
	if *tnAddr == "" || *lnAddr == "" {
		fmt.Println("Address not specified, please see usage")
		os.Exit(0)
	}

	rtm := NewRTMaster(*lnAddr, *tnAddr)
	rtm.Start()
}

// RTMaster is the Master side of reverse tunnel.
type RTMaster struct {
	lnAddr *net.TCPAddr
	tnAddr *net.TCPAddr

	cliCount int
	cliConns map[int]net.Conn
	tnConn   *net.TCPConn

	ch chan TunnelData
}

// NewRTMaster returns a RTMaster.
func NewRTMaster(lnAddr, tnAddr string) *RTMaster {
	l, err := net.ResolveTCPAddr("tcp", lnAddr)
	if err != nil {
		fmt.Println("Invilid address")
	}
	t, err := net.ResolveTCPAddr("tcp", tnAddr)
	if err != nil {
		fmt.Println("Invilid address")
	}
	return &RTMaster{
		lnAddr:   l,
		tnAddr:   t,
		cliCount: 0,
		cliConns: make(map[int]net.Conn),
		ch:       make(chan TunnelData),
	}
}

// Start start the reverse tunnel.
func (rtm *RTMaster) Start() {
	// start tunnel
	go func() {
		tn, err := net.ListenTCP("tcp", rtm.tnAddr)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		fmt.Println("Listening on:", rtm.tnAddr.String())
		rtm.tnConn, _ = tn.AcceptTCP()
		rtm.tnConn.SetKeepAlive(true)
		fmt.Println("[+] Tunnel established")
		rtm.handleTunnelConn()
	}()

	// listen externel requests
	ln, err := net.ListenTCP("tcp", rtm.lnAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	fmt.Println("Listening on:", rtm.lnAddr.String())

	for {
		rtm.cliConns[rtm.cliCount], _ = ln.Accept()
		go rtm.handleClientConn(rtm.cliCount)
		rtm.cliCount++
	}
}

func (rtm *RTMaster) handleTunnelConn() {
	// forward req to tunnel
	go func() {
		for td := range rtm.ch {
			tmp := []byte{byte(td.id), byte(td.size)}
			tmp = append(tmp, td.data[:]...)
			rtm.tnConn.Write(tmp)
			fmt.Println("[*] forward", td.size, "bytes from client", td.id)
		}
	}()

	// read from tunnel and send back to client
	for {
		var buf [bufSize]byte
		_, err := rtm.tnConn.Read(buf[:])
		if err != nil {
			return
		}
		td := TunnelData{
			id:   int(buf[0]),
			size: int(buf[1]),
			data: buf[2:],
		}
		fmt.Println("[*] send back", td.size, "bytes", "to client", td.id)
		rtm.cliConns[td.id].Write(td.data[:td.size])
	}
}

// read req and forward to tunnel
func (rtm *RTMaster) handleClientConn(id int) {
	fmt.Println("[!] Client", id, "comes ("+rtm.cliConns[id].RemoteAddr().String()+")")

	for {
		// reserve place for id and size
		var buf [bufSize - 2]byte
		n, err := rtm.cliConns[id].Read(buf[:])
		if err != nil {
			return
		}
		td := TunnelData{
			id:   id,
			size: n,
			data: buf[:],
		}
		rtm.ch <- td
	}
}
