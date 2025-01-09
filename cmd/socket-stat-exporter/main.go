package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vishvananda/netlink"
)

const metricsNamespace = "socket_stat"

type SocketStatCollector struct {
	netlinkHandle       *netlink.Handle
	bytesSent           *prometheus.GaugeVec
	sourcePortToMonitor uint16
}

// NewSocketStatCollector initializes and returns a new SocketStatCollector
func NewSocketStatCollector(sport uint16, handle *netlink.Handle) *SocketStatCollector {
	return &SocketStatCollector{
		netlinkHandle:       handle,
		sourcePortToMonitor: sport,
		bytesSent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "bytes_sent",
			Help:      "Total number of bytes sent through connection",
		}, []string{"source", "destination"}),
	}
}

// Describe sends the descriptor of each metric over to the provided channel.
func (c *SocketStatCollector) Describe(ch chan<- *prometheus.Desc) {}

func (c *SocketStatCollector) updateMetrics() {
	tcpSocketsInformation, err := c.netlinkHandle.SocketDiagTCPInfo(netlink.INET_DIAG_INFO)
	if err != nil {
		log.Println("unable to lookup tcp sockets information:", err)
		return
	}

	log.Printf("%d sockets detected\n", len(tcpSocketsInformation))
	for _, socketInfo := range tcpSocketsInformation {
		if socketInfo.InetDiagMsg.State != netlink.TCP_ESTABLISHED {
			continue
		}

		if socketInfo.InetDiagMsg.ID.SourcePort != c.sourcePortToMonitor {
			continue
		}

		info := socketInfo.TCPInfo
		if info != nil {
			socketID := socketInfo.InetDiagMsg.ID
			source := fmt.Sprintf("%s:%d", socketID.Source.String(), socketID.SourcePort)
			destination := fmt.Sprintf("%s:%d", socketID.Destination.String(), socketID.DestinationPort)

			bytesSent := info.Bytes_sent
			c.bytesSent.WithLabelValues(source, destination).Set(float64(bytesSent))
		}
	}
}

// Collect fetches the metrics and sends them over to the provided channel.
func (c *SocketStatCollector) Collect(ch chan<- prometheus.Metric) {
	c.bytesSent.Reset()

	startTime := time.Now()
	c.updateMetrics()
	duration := time.Since(startTime)
	log.Printf("metrics collected in %s\n", duration)

	c.bytesSent.Collect(ch)
}

func main() {

	sourcePort := flag.Int("sport", 443, "source port to monitor")
	prometheusAddress := flag.String("listen", "127.0.0.1:9993", "prometheus listen address")
	flag.Parse()

	netlinkHandle, err := netlink.NewHandle()
	if err != nil {
		log.Fatalln("unable to create netlink handle:", err)
	}

	collector := NewSocketStatCollector(uint16(*sourcePort), netlinkHandle)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	log.Println("Starting server on", *prometheusAddress)
	log.Fatal(http.ListenAndServe(*prometheusAddress, nil))

}
