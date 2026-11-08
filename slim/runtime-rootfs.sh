#!/bin/sh
# Build a fresh filesystem containing only the existing runtime tools and their libraries.
set -eu
mkdir -p /rootfs/bin /rootfs/usr/bin /rootfs/usr/sbin /rootfs/sbin \
    /rootfs/etc/ssl/certs /rootfs/etc/mihomo /rootfs/usr/share/zoneinfo/Europe \
    /rootfs/tmp /rootfs/run /rootfs/dev /rootfs/proc /rootfs/sys
chmod 1777 /rootfs/tmp
copy_file() {
    mkdir -p "/rootfs$(dirname "$1")"
    cp -a "$1" "/rootfs$1"
    if [ -L "$1" ]; then
        resolved=$(readlink -f "$1")
        mkdir -p "/rootfs$(dirname "$resolved")"
        cp -a "$resolved" "/rootfs$resolved"
    fi
}
for executable in /bin/busybox /sbin/tini /usr/sbin/nft /usr/bin/envsubst; do
    copy_file "$executable"
    for library in $(ldd "$executable" | awk '{for(i=1;i<=NF;i++) if($i ~ /^\//) print $i}'); do
        copy_file "$library"
    done
done
for applet in $(/bin/busybox --list); do
    if [ "$applet" != busybox ]; then ln -s busybox "/rootfs/bin/$applet"; fi
done
copy_file /usr/bin/mihomo
copy_file /entrypoint.sh
cp -a /etc/mihomo_preset /rootfs/etc/
cp /etc/ssl/certs/ca-certificates.crt /rootfs/etc/ssl/certs/
cp /usr/share/zoneinfo/Europe/Moscow /rootfs/usr/share/zoneinfo/Europe/
cp /etc/passwd /etc/group /rootfs/etc/
