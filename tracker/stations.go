// Copyright (c) 2016, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tracker

import "io"

// A Station struct represents an IEEE 802.11 access point.
type Station struct {
	SSID         string
	KnownDevices *Set
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
