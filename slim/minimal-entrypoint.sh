#!/bin/sh
set -eu
if [ "${IPTABLES:-0}" != 0 ] || ! grep -q '^nf_tables ' /proc/modules; then
    echo 'This image requires Linux nftables and TPROXY; iptables/TUN fallback is not included.' >&2
    exit 1
fi
nft list tables >/dev/null
# No /etc/alpine-release: the retained entrypoint skips dynamic package management.
exec /entrypoint.sh "$@"
