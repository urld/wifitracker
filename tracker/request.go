package tracker

import (
	"encoding/json"
	"time"
)

// JSONTime is a wrapper around time.Time to enable JSON Un-/Marshalling.

// A Request struct represents a captured IEEE 802.11 probe request.
type Request struct {
	SourceMac      string    `json:"source_mac"`
	CaptureDts     time.Time `json:"capture_dts"`
	TargetSsid     string    `json:"target_ssid"`
	SignalStrength int       `json:"signal_strength"`
}

func parseRequest(requestJSON []byte) (Request, error) {
	var request Request
	err := json.Unmarshal(requestJSON, &request)
	return request, err
}
