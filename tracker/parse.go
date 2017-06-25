// Copyright (c) 2016, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tracker

import (
	"bufio"
	"io"
	"runtime"
	"sync"

	"github.com/urld/wifitracker"
)

const bufferFactor int = 1000

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

func parseRequestJSONs(in <-chan []byte) <-chan *wifitracker.Request {
	out := make(chan *wifitracker.Request, bufferFactor)

	go func() {
		defer close(out)
		for requestJSON := range in {
			request, err := wifitracker.ParseRequest(requestJSON)
			if err != nil {
				// ignore erroneus requests
				continue
			}
			out <- &request
		}

	}()
	return out
}

func merge(cs ...<-chan *wifitracker.Request) <-chan *wifitracker.Request {
	var wg sync.WaitGroup
	out := make(chan *wifitracker.Request, bufferFactor)
	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan *wifitracker.Request) {
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

func ParseRequests(input io.Reader) <-chan *wifitracker.Request {
	requestJSONs := readRequestJSONs(input)
	var requestParsers []<-chan *wifitracker.Request
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		requests := parseRequestJSONs(requestJSONs)
		requestParsers = append(requestParsers, requests)
	}
	return merge(requestParsers...)
}
