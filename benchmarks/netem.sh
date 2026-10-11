#!/bin/sh
set -eu

usage() {
	echo "usage: $0 apply|clear INTERFACE [DELAY] [LOSS]" >&2
	exit 2
}

[ "$#" -ge 2 ] || usage
ACTION=$1
IFACE=$2
DELAY=${3:-50ms}
LOSS=${4:-0.5%}

case "$ACTION" in
apply)
	command -v tc >/dev/null 2>&1 || { echo "tc is required" >&2; exit 1; }
	exec tc qdisc replace dev "$IFACE" root netem delay "$DELAY" loss "$LOSS"
	;;
clear)
	command -v tc >/dev/null 2>&1 || { echo "tc is required" >&2; exit 1; }
	exec tc qdisc del dev "$IFACE" root
	;;
*)
	usage
	;;
esac
