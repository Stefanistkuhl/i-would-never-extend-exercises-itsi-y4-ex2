package capture

import (
	"fmt"
	"time"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type CaptureStats struct {
	TotalPackets         int            `json:"total_packets"`
	ProtocolDistribution map[string]int `json:"protocol_distribution"`
	TopSrcIPs            map[string]int `json:"top_src_ips"`
	TopDstIPs            map[string]int `json:"top_dst_ips"`
	TopPorts             map[uint16]int `json:"top_ports"`
	PacketRate           float64        `json:"packet_rate"`
	AvgPacketSize        float64        `json:"avg_packet_size"`
	TLSVersions          map[string]int `json:"tls_versions"`
	DNSQueries           int            `json:"dns_queries"`
	DurationSeconds      int64          `json:"duration_seconds"`
	FirstPacketTime      time.Time      `json:"first_packet_time"`
	LastPacketTime       time.Time      `json:"last_packet_time"`
}

func AnalyzeCaptureFile(cfg config.Config, filePath string) (CaptureStats, error) {
	handle, err := pcap.OpenOffline(filePath)
	if err != nil {
		return CaptureStats{}, fmt.Errorf("failed to open pcap: %w", err)
	}
	defer handle.Close()

	stats := CaptureStats{
		ProtocolDistribution: make(map[string]int),
		TopSrcIPs:            make(map[string]int),
		TopDstIPs:            make(map[string]int),
		TopPorts:             make(map[uint16]int),
		TLSVersions:          make(map[string]int),
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	var totalPackets int
	var totalBytes int64
	var firstTime, lastTime time.Time

	for packet := range packetSource.Packets() {
		if firstTime.IsZero() {
			firstTime = packet.Metadata().Timestamp
		}
		lastTime = packet.Metadata().Timestamp

		totalPackets++
		totalBytes += int64(len(packet.Data()))

		if ipv4Layer := packet.Layer(layers.LayerTypeIPv4); ipv4Layer != nil {
			ipv4 := ipv4Layer.(*layers.IPv4)
			stats.ProtocolDistribution["IPv4"]++
			stats.TopSrcIPs[ipv4.SrcIP.String()]++
			stats.TopDstIPs[ipv4.DstIP.String()]++
		}

		if ipv6Layer := packet.Layer(layers.LayerTypeIPv6); ipv6Layer != nil {
			ipv6 := ipv6Layer.(*layers.IPv6)
			stats.ProtocolDistribution["IPv6"]++
			stats.TopSrcIPs[ipv6.SrcIP.String()]++
			stats.TopDstIPs[ipv6.DstIP.String()]++
		}

		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp := tcpLayer.(*layers.TCP)
			stats.ProtocolDistribution["TCP"]++
			stats.TopPorts[uint16(tcp.DstPort)]++

			if tcp.DstPort == 443 {
				stats.TLSVersions[detectTLSVersion(packet)]++
			}
		}

		if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			udp := udpLayer.(*layers.UDP)
			stats.ProtocolDistribution["UDP"]++
			stats.TopPorts[uint16(udp.DstPort)]++

			if udp.DstPort == 53 {
				if dnsLayer := packet.Layer(layers.LayerTypeDNS); dnsLayer != nil {
					dns := dnsLayer.(*layers.DNS)
					if !dns.QR {
						stats.DNSQueries++
					}
				}
			}
		}

		if packet.Layer(layers.LayerTypeICMPv4) != nil {
			stats.ProtocolDistribution["ICMP"]++
		}
		if packet.Layer(layers.LayerTypeICMPv6) != nil {
			stats.ProtocolDistribution["ICMPv6"]++
		}
	}

	stats.FirstPacketTime = firstTime
	stats.LastPacketTime = lastTime
	stats.DurationSeconds = int64(lastTime.Sub(firstTime).Seconds())
	stats.TotalPackets = totalPackets

	if totalPackets > 0 {
		stats.AvgPacketSize = float64(totalBytes) / float64(totalPackets)
	}

	if stats.DurationSeconds > 0 {
		stats.PacketRate = float64(totalPackets) / float64(stats.DurationSeconds)
	}

	return stats, nil
}

func detectTLSVersion(packet gopacket.Packet) string {
	appLayer := packet.ApplicationLayer()
	if appLayer == nil {
		return "Unknown"
	}

	payload := appLayer.Payload()
	if len(payload) < 5 || payload[0] != 0x16 {
		return "Unknown"
	}

	version := (int(payload[1]) << 8) | int(payload[2])
	switch version {
	case 0x0301:
		return "TLS1.0"
	case 0x0302:
		return "TLS1.1"
	case 0x0303:
		return "TLS1.2"
	case 0x0304:
		return "TLS1.3"
	default:
		return "Unknown"
	}
}
