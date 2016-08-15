#!/bin/sh

case "$1" in
    "start")
        service network-manager stop
        ifconfig $2 down
        iwconfig $2 mode monitor
        ifconfig $2 up
        ;;
    "stop")
        ifconfig $2 down
        iwconfig $2 mode managed
        ifconfig $2 up
        service network-manager start
        ;;
    *)
        echo "usage: monitor.sh start|stop <iface>"
        ;;
esac
