
[![Build Status](https://travis-ci.org/urld/wifitracker.svg?branch=master)](https://travis-ci.org/urld/wifitracker)
[![Go Report Card](https://goreportcard.com/badge/github.com/urld/wifitracker)](https://goreportcard.com/report/github.com/urld/wifitracker)
[![GoDoc](https://godoc.org/github.com/urld/wifitracker?status.svg)](https://godoc.org/github.com/urld/wifitracker)

This is a golang port of [py-wifi-tracker](https://github.com/urld/py-wifi-tracker).

## Usage

Record probe requests:
```
$ ./monitor.sh start wlan0
$ wifisniff mon0 > requests.log
```

Analyze recored probe requests:
```
$ wifianalyze devices < requests.log
```

## Install

```
apt install libpcap0.8 libpcap0.8-dev
go get github.com/urld/wifitracker
```

