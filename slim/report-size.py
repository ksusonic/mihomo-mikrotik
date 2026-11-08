"""Measure the exported filesystem, excluding tar padding and image metadata."""
import json
import sys
import tarfile

with tarfile.open(sys.argv[1]) as rootfs:
    entries = [entry for entry in rootfs if entry.isfile()]
    binary = next(entry for entry in entries if entry.name.lstrip('./') == 'usr/bin/mihomo')
    print(json.dumps({
        'variant': sys.argv[2],
        'rootfs_bytes': sum(entry.size for entry in entries),
        'mihomo_bytes': binary.size,
    }, indent=2))
