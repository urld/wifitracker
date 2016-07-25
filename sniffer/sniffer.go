package sniffer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/durl/go-wifi-tracker/tracker"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// A ProbeRequest struct represents a captured IEEE 802.11 probe request packet.
type ProbeRequest struct {
	MAC            string
	SSID           string
	RSSI           int8
	VendorSpecific []byte
	CaptureDTS     time.Time
}

func (pr *ProbeRequest) decodeProbeRequestLayer(probeLayer *layers.Dot11MgmtProbeReq) {
	var body []byte
	body = probeLayer.LayerContents()
	for i := uint64(0); i < uint64(len(body)); {
		id := layers.Dot11InformationElementID(body[i])
		i++
		switch id {
		case layers.Dot11InformationElementIDSSID:
			elemLen := uint64(body[i])
			i++
			if elemLen > 0 {
				pr.SSID = string(body[i : i+elemLen])
				i += elemLen
			}
			break
		case layers.Dot11InformationElementIDVendor:
			pr.VendorSpecific = body[i+1:]
			return
		default:
			elemLen := uint64(body[i])
			i += 1 + elemLen
			break
		}
	}
}

func handlePacket(packet gopacket.Packet) {
	probeRequest := ProbeRequest{CaptureDTS: time.Now()}
	if l1 := packet.Layer(layers.LayerTypeDot11); l1 != nil {
		dot11, _ := l1.(*layers.Dot11)
		probeRequest.MAC = dot11.Address2.String()
		if l2 := packet.Layer(layers.LayerTypeDot11MgmtProbeReq); l2 != nil {
			dot11p, _ := l2.(*layers.Dot11MgmtProbeReq)
			probeRequest.decodeProbeRequestLayer(dot11p)
			if l1 := packet.Layer(layers.LayerTypeRadioTap); l1 != nil {
				dot11r, _ := l1.(*layers.RadioTap)
				probeRequest.RSSI = dot11r.DBMAntennaSignal
			}
			rq := tracker.Request{CaptureDts: probeRequest.CaptureDTS, SignalStrength: 0, SourceMac: probeRequest.MAC, TargetSsid: probeRequest.SSID}
			b, _ := json.Marshal(rq)
			// print request to stdout:
			fmt.Println(string(b))
		}
	}
}

func Sniff(iface string) {
	handle, err := pcap.OpenLive(iface, 1600, true, 0)
	if err != nil {
		panic(err)
	}
	// only capture probe request packets
	err = handle.SetBPFFilter("type mgt subtype probe-req")
	if err != nil {
		panic(err)
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		handlePacket(packet)
	}
}
