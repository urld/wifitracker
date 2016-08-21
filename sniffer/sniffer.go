// Copyright (c) 2016, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sniffer

import (
	"time"

	"github.com/durl/wifitracker"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// A CapturedRequest struct represents a captured IEEE 802.11 probe request packet.
type capturedRequest struct {
	MAC            string
	SSID           string
	RSSI           int8
	VendorSpecific []byte
	CaptureDTS     time.Time
}

func (pr *capturedRequest) decodeProbeRequestLayer(probeLayer *layers.Dot11MgmtProbeReq) {
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

func handlePacket(packet gopacket.Packet, out chan<- wifitracker.Request) {
	probeRequest := capturedRequest{CaptureDTS: time.Now()}
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
			rq := wifitracker.Request{CaptureDts: probeRequest.CaptureDTS, SignalStrength: 0, SourceMac: probeRequest.MAC, TargetSsid: probeRequest.SSID}
			out <- rq
		}
	}
}

// Setup creates a new handle for the given interface.
// The handle is configured to capture only IEEE 802.11 probe request packets.
func Setup(iface string) (*pcap.Handle, error) {
	handle, err := pcap.OpenLive(iface, 1600, true, 0)
	if err != nil {
		return handle, err
	}
	// only capture probe request packets
	err = handle.SetBPFFilter("type mgt subtype probe-req")
	if err != nil {
		return handle, err
	}
	return handle, nil
}

// Sniff captures probe requests using the given handle and writes them to a channel.
func Sniff(handle *pcap.Handle) <-chan wifitracker.Request {
	out := make(chan wifitracker.Request, 0)

	go func() {
		defer close(out)
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			handlePacket(packet, out)
		}
	}()
	return out
}
