package main

import (
	"flag"
	"log"
	"net"
	"time"

	"github.com/vishvananda/netlink"
)

var tcpStateMap = map[uint8]string{
	netlink.TCP_ESTABLISHED: "ESTABLISHED",
	netlink.TCP_SYN_SENT:    "SYN_SENT",
	netlink.TCP_SYN_RECV:    "SYN_RECV",
	netlink.TCP_FIN_WAIT1:   "FIN_WAIT1",
	netlink.TCP_FIN_WAIT2:   "FIN_WAIT2",
	netlink.TCP_TIME_WAIT:   "TIME_WAIT",
	netlink.TCP_CLOSE:       "CLOSE",
	netlink.TCP_CLOSE_WAIT:  "CLOSE_WAIT",
	netlink.TCP_LAST_ACK:    "LAST_ACK",
	netlink.TCP_LISTEN:      "LISTEN",
	netlink.TCP_CLOSING:     "CLOSING",
	netlink.TCP_NEW_SYN_REC: "NEW_SYN_REC",
	netlink.TCP_MAX_STATES:  "MAX_STATES",
}

func main() {

	localAddress := flag.String("local-addr", "127.0.0.1:59647", "source socket address")
	remoteAddress := flag.String("remote-addr", "127.0.0.1:8080", "destination socket address")
	interval := flag.Duration("interval", 1*time.Second, "socket scrape interval")
	flag.Parse()

	h, err := netlink.NewHandle()
	if err != nil {
		log.Fatalln("unable to create netlink handle:", err)
	}
	defer h.Close()

	localAddr, err := net.ResolveTCPAddr("tcp", *localAddress)
	remoteAddr, err := net.ResolveTCPAddr("tcp", *remoteAddress)

	socket, err := netlink.SocketGet(localAddr, remoteAddr)
	if err != nil {
		log.Fatalf("socket not found. local %s, remote %s\n", localAddr, remoteAddr)
	}

	lastValue := uint64(0)
	for {
		ss, err := h.SocketDiagTCPInfoSingle(netlink.INET_DIAG_INFO, socket.ID)
		if err != nil {
			log.Fatalln("tcp info fetch failed:", err)
		}

		if ss.TCPInfo == nil {
			log.Println("socket gone")
			return
		}

		if lastValue != 0 {
			diff := ss.TCPInfo.Bytes_sent - lastValue
			if diff > 0 {
				log.Println(diff / 1024)
			}
		}

		lastValue = ss.TCPInfo.Bytes_sent
		time.Sleep(*interval)
	}
}
