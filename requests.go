// Copyright (c) 2016, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wifitracker

import (
	"encoding/json"
	"time"
)

// A Request struct represents a captured IEEE 802.11 probe request.
type Request struct {
	SourceMac      string    `json:"source_mac"`
	CaptureDts     time.Time `json:"capture_dts"`
	TargetSsid     string    `json:"target_ssid"`
	SignalStrength int       `json:"signal_strength"`
}

// ParseRequest parses a request in json format.
func ParseRequest(requestJSON []byte) (Request, error) {
	var request Request
	err := json.Unmarshal(requestJSON, &request)
	return request, err
}
