package main

import (
	"bufio"
	"encoding/json"
	"os"
	"time"
	"fmt"
)

const jsonTimeFmt string = "2006-01-02 15:04:05.000000"

//Extension for time.Time to get unmarshaled from JSON
type JsonTime struct {
	time.Time
}

// Method to unmarshal time.Time from json.
func (t *JsonTime) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	t.Time, err = time.Parse(jsonTimeFmt, string(b))
	return
}

// Method to marsjal time.Time to json.
func (t *JsonTime) MarshalJSON() ([]byte, error) {
	return []byte(t.Time.Format(jsonTimeFmt)), nil
}

type Request struct {
	SourceMac      string   `json:"source_mac"`
	CaptureDts     JsonTime `json:"capture_dts"`
	TargetSsid     string   `json:"target_ssid"`
	SignalStrength int      `json:"signal_strength"`
}

func parseRequest(requestJson []byte) (Request, []byte) {
	var request Request
	err := json.Unmarshal(requestJson, &request)
	if err != nil {
		return request, requestJson
	}
	return request, nil
}

type Device struct {
	DeviceMac     string
	Alias         string
	KnownSsids    Set
	LastSeenDts   JsonTime
	VendorCompany string
	VendorCountry string
}

func (d *Device) AddSsid (ssid string){
	d.KnownSsids.Add(ssid)
}

type Set struct {
	 set map[string]bool
}

func (s *Set) Add(element string) {
	if s.set == nil{
		s.set = make(map[string]bool)
	}
	s.set[element] = true
}

func readRequestJsons(requestFilePath string) <-chan []byte {
	c := make(chan []byte)


	f, _ := os.Open(requestFilePath)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	go func (){
		for scanner.Scan() {
			c <- scanner.Bytes()
		}
	}()
	return c
}

func main() {
	requestFilePath := "/var/opt/wifi-tracker/requests"

	var corruptLines [][]byte
	devices := make(map[string]Device)
	f, _ := os.Open(requestFilePath)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		request, corrupt := parseRequest(scanner.Bytes())
		if corrupt != nil && len(corrupt) != 0 {

			corruptLines = append(corruptLines, corrupt)
			continue
		}
		if device, exists := devices[request.SourceMac]; exists {
			device.LastSeenDts = request.CaptureDts
			device.AddSsid(request.TargetSsid)
		}

		device := Device{
			DeviceMac: request.SourceMac,
			LastSeenDts: request.CaptureDts,
			KnownSsids: Set{},
		}
		device.AddSsid(request.TargetSsid)
		devices[request.SourceMac] = device
	}

	fmt.Println(len(devices))
	for _, d := range devices {
		fmt.Println(d)
	}
}
