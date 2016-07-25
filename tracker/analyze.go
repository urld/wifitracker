package tracker

import (
	"bufio"
	"encoding/json"
	"io"
	"runtime"
	"sync"
	"time"
)

const bufferFactor int = 1000

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

// A Device struct represents a IEEE 802.11 device which was actively scanning for access points.
type Device struct {
	DeviceMac     string
	Alias         *string
	KnownSsids    *Set
	LastSeenDts   time.Time
	VendorCompany *string
	VendorCountry *string
}

// A Station struct represents an IEEE 802.11 access point.
type Station struct {
	SSID         string
	KnownDevices *Set
}

func readRequestJSONs(input io.Reader) <-chan []byte {
	out := make(chan []byte, bufferFactor)

	go func() {
		defer close(out)
		scanner := bufio.NewScanner(input)
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
				// ignore erroneus requests
				continue
			}
			out <- &request
		}

	}()
	return out
}

func merge(cs ...<-chan *Request) <-chan *Request {
	var wg sync.WaitGroup
	out := make(chan *Request, bufferFactor)
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

func ParseRequests(input io.Reader) <-chan *Request {
	requestJSONs := readRequestJSONs(input)
	var requestParsers []<-chan *Request
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		requests := parseRequestJSONs(requestJSONs)
		requestParsers = append(requestParsers, requests)
	}
	return merge(requestParsers...)
}

func AggregateStations(input io.Reader) map[string]interface{} {
	in := ParseRequests(input)
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

func AggregateDevices(input io.Reader) map[string]interface{} {
	in := ParseRequests(input)
	devices := make(map[string]interface{})
	for request := range in {
		// check if device was already identified:
		if deviceI, exists := devices[request.SourceMac]; exists {
			device, _ := deviceI.(Device)
			// update LastSeenDts:
			if device.LastSeenDts.Before(request.CaptureDts) {
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
