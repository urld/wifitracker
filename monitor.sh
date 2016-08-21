#!/bin/sh

usage(){
        echo "usage: monitor.sh start|stop <iface>"
        exit 2
}

if [ "$#" -ne 2 ]; then
    usage
fi

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
        usage
        ;;
esac

