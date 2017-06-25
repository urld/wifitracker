// Copyright (c) 2016, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/urld/wifitracker/tracker"
)

const usage = "usage: wifianalyze devices|stations"

func printEntities(entities map[string]interface{}) {
	for _, entity := range entities {

		entityJSON, err := json.MarshalIndent(entity, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(entityJSON))
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println(usage)
		os.Exit(2)
	}
	typeArg := os.Args[1]

	input := bufio.NewReader(os.Stdin)

	switch typeArg {
	case "devices":
		devices := tracker.AggregateDevices(input)
		printEntities(devices)
	case "stations":
		stations := tracker.AggregateStations(input)
		printEntities(stations)
	default:
		fmt.Println(usage)
		os.Exit(2)
	}
}
