package tracker

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
		f, err := os.Open(requestFilePath)
		if err != nil {
			return
		}
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

func merge(cs ...<-chan *Request) <-chan *Request {
	var wg sync.WaitGroup
	out := make(chan *Request, bufferFactor*runtime.NumCPU())

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan *Request) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func ParseRequests(requestFilePath string) <-chan *Request {
	requestJSONs := readRequestJSONs(requestFilePath)
	var requestParsers []<-chan *Request
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		requests := parseRequestJSONs(requestJSONs)
		requestParsers = append(requestParsers, requests)
	}
	return merge(requestParsers...)
}

func AggregateStations(requestFilePath string) map[string]interface{} {
	in := ParseRequests(requestFilePath)
	stations := make(map[string]interface{})
	for request := range in {
		if request.TargetSsid == "" {
			// unable to identify station
			continue
		}

		// check if station was already identified:
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
	}
	return stations
}

func AggregateDevices(requestFilePath string) map[string]interface{} {
	in := ParseRequests(requestFilePath)
	devices := make(map[string]interface{})
	for request := range in {
		// check if device was already identified:
		if deviceI, exists := devices[request.SourceMac]; exists {
			device, _ := deviceI.(Device)
			// update LastSeenDts:
			if device.LastSeenDts.Time.Before(request.CaptureDts.Time) {
				device.LastSeenDts = request.CaptureDts
			}
			// update KnownSsids:
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
	}
	return devices
}
