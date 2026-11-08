# ARM64 TPROXY builds

This fork keeps the upstream container tools and protocol support. It builds
Mihomo v1.19.33 with Go 1.26.10 and the installed image's compatibility patches.
No local router configuration, provider credentials or subscriptions belong here.

## Variants

- `baseline`: gVisor enabled, normal Go inlining, for comparison.
- `slim`: no `with_gvisor` tag; suitable for Linux TPROXY, not gVisor TUN.
- `compact`: same as slim, with `-gcflags=all=-l`. This disables compiler
  inlining to reduce executable size without deleting protocols. Benchmark it
  before deployment: a smaller binary can cost forwarding performance.

Run **ARM64 slim build** manually under GitHub Actions. Each variant produces a
Docker-save `.tar.gz` accepted by RouterOS, a SHA256 checksum, binary version,
local-forwarding benchmark, and unpacked filesystem size report. Nothing is
published to a registry or installed automatically. Fork builds no longer run
an inherited daily release/publish schedule.

The public module proxy and checksum database apply only to the build stage.
Local Go commands must use the shared build cache; do not change global Go or Git
settings. On GitHub-hosted runners, BuildKit caches the Go build directory.

## Pinned provenance

- Container source: `wiktorbgu/mihomo-mikrotik` commit
  `49637af0119e04d181c048d633f98664ce690999`.
- Mihomo source: `3f27ef39f8c96bbfbc8b0b9bb10f5c9ef60149d8` (`v1.19.33`).
- REALITY key-share patch: upstream commit
  `04653b5bcaec342a0573edb86a85e515869c88ce` (PR #3326), vendored under patches.
- Existing provider tolerance and REALITY client version compatibility edits
  are retained in Dockerfile.slim, with assertions that they applied.
- `entrypoint.sh` is taken from the installed ARM64 OCI image manifest
  `sha256:f04e9012e5ed65e65f63543b9b00a9124d7c22634a1283104112fa2e8430e9b4`;
  SHA256 `d445f255df05d877bb30b7d5b1eb6fae13fa93ee17e8800597606f3c29243d9f`.
  It is deliberately retained instead of substituting a different upstream
  revision's network setup code.
- Build and runtime base image digests are pinned in Dockerfile.slim.

## Deployment gate

Require at least 10 MiB less unpacked filesystem space than the installed
container, successful parsing of the actual private configuration, and working
VLESS REALITY. Retain the current container if neither variant meets this gate.

Before replacement, export the stopped working image and its mounted directory
to a private off-router backup and verify the archive/checksum. Use a temporary
RAM disk for import; internal flash cannot hold both images at once. Stop/remove
the old container and import the candidate with the same veth, mounts, environment,
limits, healthcheck and restart policy. Do not enable automatic image updates.

After import verify transparent TCP/UDP, Grafana remote-write, inaccessible LAN
proxy/API ports, blocked direct fallback, and automatic recovery from a paused
process. Revert to the saved image if any check fails. Existing private mounted
health/security scripts continue to apply. The only retained controller is the
loopback health endpoint; no UI or public proxy is added by the slim build.
