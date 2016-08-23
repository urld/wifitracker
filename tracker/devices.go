// Copyright (c) 2016, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tracker

import (
	"io"
	"time"
)

// A Device struct represents a IEEE 802.11 device which was actively scanning for access points.
type Device struct {
	DeviceMac     string
	Alias         *string
	KnownSsids    *Set
	LastSeenDts   time.Time
	VendorCompany *string
	VendorCountry *string
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
