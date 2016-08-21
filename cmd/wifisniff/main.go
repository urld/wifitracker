// Copyright (c) 2016, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/durl/wifitracker"
	"github.com/durl/wifitracker/sniffer"
)

const usage = "usage: wifisniff <interface>"

func exit(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

func exitUsage() {
	fmt.Fprintln(os.Stderr, usage)
	os.Exit(2)
}

func main() {

	if len(os.Args) != 2 {
		exitUsage()
	}
	iface := os.Args[1]

	handle, err := sniffer.Setup(iface)
	if err != nil {
		exit(err)
	}
	requests := sniffer.Sniff(handle)

	printRequests(requests)
}

func printRequests(rqs <-chan wifitracker.Request) {
	for rq := range rqs {
		rqJson, _ := json.Marshal(rq)
		// ignore marshalling errors
		fmt.Println(string(rqJson))
	}
}
