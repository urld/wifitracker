package main

import (
	"bufio"
	"encoding/json"
	"os"
	"runtime"
	"sync"
	"time"
)

const jsonTimeFmt string = "2006-01-02 15:04:05.000000"
const bufferFactor int = 100

// A Set is a unordered collection of unique string elements.
type Set struct {
	set map[string]bool
}

// Add an element to the Set.
func (s *Set) Add(element string) {
	if element != "" {
		if s.set == nil {
			s.set = make(map[string]bool)
		}
		s.set[element] = true
	}
}

// MarshalJSON returns the JSON encoding of the Set.
// The Set is converted to a JSON Array, and empty strings are ignored.
func (s *Set) MarshalJSON() ([]byte, error) {
	elements := make([]string, 0, len(s.set))
	for element := range s.set {
		if element != "" {
			elements = append(elements, element)
		}
	}
	setJSON, err := json.Marshal(elements)
	return setJSON, err
}

// JSONTime is a wrapper around time.Time to enable JSON Un-/Marshalling.
type JSONTime struct {
	time.Time
}

// UnmarshalJSON parses the JSON-encoded datetime stamp.
func (t *JSONTime) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	t.Time, err = time.Parse(jsonTimeFmt, string(b))
	return
}

// MarshalJSON returns the JSON encoding of the JSONTime value.
func (t *JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(t.Time.Format(jsonTimeFmt)), nil
}

// A Request struct represents a captured IEEE 802.11 probe request.
type Request struct {
	SourceMac      string   `json:"source_mac"`
	CaptureDts     JSONTime `json:"capture_dts"`
	TargetSsid     string   `json:"target_ssid"`
	SignalStrength int      `json:"signal_strength"`
}

func parseRequest(requestJSON []byte) (Request, error) {
	var request Request
	err := json.Unmarshal(requestJSON, &request)
	return request, err
}

// A Device struct represents a IEEE 802.11 device which was actively scanning for access points.
type Device struct {
	DeviceMac     string
	Alias         *string
	KnownSsids    *Set
	LastSeenDts   JSONTime
	VendorCompany *string
	VendorCountry *string
}

// A Station struct represents an IEEE 802.11 access point.
type Station struct {
	SSID         string
	KnownDevices *Set
}

func readRequestJSONs(requestFilePath string) <-chan []byte {
	out := make(chan []byte, runtime.NumCPU()*100)

	go func() {
		f, _ := os.Open(requestFilePath)
		defer f.Close()
		defer close(out)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			// copy scan result because it may get overwritten by the next scan result:
			var line []byte
			line = append(line, scanner.Bytes()...)
			out <- line
		}
	}()
	return out
}

func parseRequestJSONs(in <-chan []byte) <-chan *Request {
	out := make(chan *Request, bufferFactor)

	go func() {
		defer close(out)
		for requestJSON := range in {
			request, err := parseRequest(requestJSON)
			if err != nil {
				continue
			}
			out <- &request
		}

	}()
	return out
}

func aggregateStations(in <-chan *Request, stations map[string]interface{}, stationsMutex *sync.Mutex) <-chan bool {
	done := make(chan bool)

	go func() {
		defer close(done)
		for request := range in {
			if request.TargetSsid == "" {
				continue
			}
			stationsMutex.Lock()
			if stationI, exists := stations[request.TargetSsid]; exists {
				station, _ := stationI.(Station)
				station.KnownDevices.Add(request.SourceMac)
			} else {
				station := Station{
					SSID:         request.TargetSsid,
					KnownDevices: &Set{},
				}
				station.KnownDevices.Add(request.SourceMac)
				stations[request.TargetSsid] = station
			}
			stationsMutex.Unlock()
		}
		done <- true
	}()
	return done
}

func aggregateDevices(in <-chan *Request, devices map[string]interface{}, devicesMutex *sync.Mutex) <-chan bool {
	done := make(chan bool)

	go func() {
		defer close(done)
		for request := range in {
			devicesMutex.Lock()
			if deviceI, exists := devices[request.SourceMac]; exists {
				device, _ := deviceI.(Device)
				if device.LastSeenDts.Time.Before(request.CaptureDts.Time) {
					device.LastSeenDts = request.CaptureDts
				}
				device.KnownSsids.Add(request.TargetSsid)
			} else {
				device := Device{
					DeviceMac:   request.SourceMac,
					LastSeenDts: request.CaptureDts,
					KnownSsids:  &Set{},
				}
				device.KnownSsids.Add(request.TargetSsid)
				devices[request.SourceMac] = device
			}
			devicesMutex.Unlock()
		}
		done <- true
	}()
	return done
}
