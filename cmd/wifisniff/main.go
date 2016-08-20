// Copyright (c) 2016, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/durl/wifitracker/sniffer"
)

const usage = "usage: sniffer <interface>"

func main() {

	if len(os.Args) != 2 {
		fmt.Println(usage)
		os.Exit(2)
	}
	iface := os.Args[1]

	sniffer.Sniff(iface)
}
