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
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"github.com/docopt/docopt-go"
)

const Version string = "0.1.0"
const requestFilePath string = "/var/opt/wifi-tracker/requests"

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
	// init runtime:
	runtime.GOMAXPROCS(runtime.NumCPU())
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
		requestJSONs := readRequestJSONs(requestFilePath)
		var requestParsers []<-chan *Request
		for i := 0; i < runtime.NumCPU(); i++ {
			requests := parseRequestJSONs(requestJSONs)
			requestParsers = append(requestParsers, requests)
		}

		entities := make(map[string]interface{})
		entitiesMutex := &sync.Mutex{}

		var done <-chan bool
		if args["devices"].(bool) {
			done = aggregateDevices(merge(requestParsers...), entities, entitiesMutex)
			<-done
		} else if args["stations"].(bool) {
			done = aggregateStations(merge(requestParsers...), entities, entitiesMutex)
			<-done
		} else if args["aliases"].(bool){
			fmt.Println("not implemented")
			os.Exit(2)
		}
		printEntities(entities)
	} else if args["sniff"].(bool){
		sniff(args["<interface>"].(string))
	} else {
		fmt.Println("not implemented")
		os.Exit(2)
	}
}
