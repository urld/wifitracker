/*
wifitracker
Copyright (C) 2015 David Url <david@x00.at>

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published
by the Free Software Foundation, version 2.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/durl/go-wifi-tracker/sniffer"
	"github.com/durl/go-wifi-tracker/tracker"
)

const Version string = "0.1.0"

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
	usage := `wifi-tracker: Track wifi devices in your area.

Usage:
    wifi-tracker sniff <interface> [options]
    wifi-tracker show (devices|stations|aliases) [<id>] [options]
    wifi-tracker set <device_mac> <alias> [--force]
    wifi-tracker kill
    wifi-tracker monitor <interface> (start|stop) [--force]
    wifi-tracker -h | --help
    wifi-tracker --version

Options:
    -h --help           Show help.
    --debug             Print debugging messages.
    --nooui             Omit OUI vendor lookup. This might be usefull if
                        no internet connection is availaible.
    --noalias           Ignore alias file.

Commands:
    sniff           Sniff probe requests sent by devices in your area.
    show            Show tracked devices or wifi stations.
                    (this operation could take some time)
    set             Set an alias for a known device.
    kill            Kill the last startet sniffer process.
    monitor         Start or stop monitor mode on specified interface.
`
	// parse arguments:
	args, _ := docopt.Parse(usage, nil, true, Version, false)

	// read arguments:
	if args["show"].(bool) {
		input := bufio.NewReader(os.Stdin)
		if args["devices"].(bool) {
			devices := tracker.AggregateDevices(input)
			printEntities(devices)
		} else if args["stations"].(bool) {
			stations := tracker.AggregateStations(input)
			printEntities(stations)
		} else if args["aliases"].(bool) {
			fmt.Println("not implemented")
			os.Exit(2)
		}
	} else if args["sniff"].(bool) {
		sniffer.Sniff(args["<interface>"].(string))
	} else {
		fmt.Println("not implemented")
		os.Exit(2)
	}
}
