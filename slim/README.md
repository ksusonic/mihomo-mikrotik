# ARM64 TPROXY builds

This fork keeps the upstream container tools and protocol support. It builds
Mihomo v1.19.33 with Go 1.26.10 and the installed image's compatibility patches.
No local router configuration, provider credentials or subscriptions belong here.

## Variants

- `baseline`: gVisor enabled, normal Go inlining, for comparison.
- `slim`: no `with_gvisor` tag; suitable for Linux TPROXY, not gVisor TUN.
- `minimal`: slim with only BusyBox, tini, nftables, envsubst, required shared
  libraries, CA trust and the Moscow timezone. No runtime package installation
  or iptables/TUN fallback.
- `minimal-compact`: minimal with disabled Go inlining.
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

## Measured v1.19.33 results

Manual build: https://github.com/ksusonic/mihomo-mikrotik/actions/runs/42477233958

| Variant | Unpacked regular file bytes | Executable bytes |
| --- | ---: | ---: |
| baseline | 69,802,294 | 57,278,590 |
| slim | 60,037,430 | 47,513,726 |
| compact | 55,646,518 | 43,122,814 |
| minimal-compact | 47,863,126 | 43,122,814 |

Minimal-compact saves 21,939,168 bytes (20.92 MiB) of unpacked files versus
baseline. Filesystem metadata/allocation may differ on RouterOS. It passed
configuration parsing, runtime tool checks, HTTP proxy smoke tests and local
forwarding benchmarks. A private ARM64 TPROXY deployment additionally passed
HTTPS and UDP STUN forwarding; three 5 MB VLESS downloads showed no observed
throughput regression (network measurements are not controlled benchmarks).

Prefer `minimal` if normal compiler inlining is required. `minimal-compact`
retains the same protocols but trades inlining for approximately 4.2 MiB more
space. These images do not implement runtime package installation or automatic
upgrades; rebuild explicitly after upstream/security updates.

## Lean feature build

The manual workflow now builds `lean`: minimal-compact plus the upstream
`no_easytier no_zerotier no_fake_tcp` tags. VLESS REALITY, TPROXY, UDP, DNS,
sniffing and rule providers remain available. These disabled transports cannot
be added later through configuration; rebuilding without LEAN_FEATURES restores
them. Existing variants remain buildable through Dockerfile.slim arguments.

Measured lean build: 35,738,966 unpacked regular-file bytes; executable
30,998,654 bytes. This saves another 12,124,160 bytes (11.56 MiB) versus
minimal-compact. Build and runtime checks:
https://github.com/ksusonic/mihomo-mikrotik/actions/runs/42478066021

Private TPROXY deployment passed HTTPS, UDP STUN and blocked direct fallback.
Actual router free flash increased from 42.2 to 53.8 MiB. Three VLESS downloads
measured 3.71, 4.71 and 4.63 MB/s; no clear regression under variable network
conditions.

## Experimental VLESS/TCP/REALITY + TPROXY profile

`VLESS_ONLY=1` applies `restrict.go` to the immutable core source before building.
The workflow currently builds this profile. Outbound parsing accepts only VLESS
and the DIRECT/DNS/REJECT/rematch helpers needed by routing. VLESS requires TCP,
TLS and a REALITY public key; WS, gRPC, XHTTP and other network modes fail explicitly.
Configured inbound listeners accept only TPROXY, and legacy HTTP/SOCKS/mixed,
TUN, Shadowsocks, VMess and TUIC settings are rejected. The matching legacy
listener constructors are disabled. VLESS alternate transport branches and their
client cleanup paths are removed before compilation so the linker can drop
unreachable implementations. Shared TLS/DNS code and UDP/XUDP remain intact.

DNS, sniffing, rule providers including community RULE-SET files, proxy groups,
loopback health API and intentional DIRECT routing are preserved. No community
list or routing policy is changed by installing this profile.

Measured filesystem: 25,515,350 bytes; executable: 20,775,038 bytes.
Previous lean filesystem: 35,738,966 bytes. Router free flash: 53.8 -> 63.8 MiB.
Validated build (including restricted parser tests):
https://github.com/ksusonic/mihomo-mikrotik/actions/runs/42479469019
Private configuration parsing, transparent HTTPS, UDP STUN and missing-route
blocking passed. Three repeated 5 MB downloads measured 3.54, 4.80 and 5.14 MB/s;
one earlier transfer attempt timed out. These are variable-network observations,
not a controlled throughput benchmark.

This is a maintained core customization, not just upstream feature flags. Review
and re-test the transformation and parser restrictions when changing the pinned
Mihomo version. Deploy the saved lean archive to restore wider protocol support.
