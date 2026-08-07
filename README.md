[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/udpscan/release.yml?branch=main&label=BUILD)](https://github.com/yylego/udpscan/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/udpscan)](https://pkg.go.dev/github.com/yylego/udpscan)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/udpscan/main.svg)](https://coveralls.io/github/yylego/udpscan?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25+-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/udpscan.svg)](https://github.com/yylego/udpscan/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/udpscan)](https://goreportcard.com/report/github.com/yylego/udpscan)

# udpscan

Detect LAN hosts via UDP broadcast.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->

## CHINESE README

[中文说明](README.zh.md)

<!-- TEMPLATE (EN) CLOSE: LANGUAGE NAVIGATION -->

## Main Features

- 🎯 **Name instead of address**: hosts announce a nickname, so machines are found by name rather than by an address nobody remembers
- ⚡ **Ready-to-run ssh command**: replies carry the login username and sshd port, and the scanner composes the command from the address the reply actually came from
- 🌐 **Every interface covered**: each local interface is probed on its own broadcast address, and point-to-point tunnels are skipped by capability rather than by name
- 📦 **Made for scripts**: `--json` emits a structured result, the exit code reports whether anything was found, and progress never pollutes stdout
- 🔒 **Firewall-friendly**: pinning the source port narrows the required rule from a whole ephemeral range down to a single port
- 🩺 **Failures stay loud**: oversized replies, invalid flags and protocol mismatches get reported rather than silently swallowed

---

## Installation

```bash
go install github.com/yylego/udpscan/cmd/udpscan@latest
```

## Usage

### Start Host Detection Service

Run on hosts you want to detect:

```bash
udpscan server --name xiaozhixiang
```

Options:
- `--name, -n` : Host nickname (required)
- `--username, -u` : Account others should use to ssh in (falls back to the current account)
- `--ssh-port` : Port the local sshd listens on (default: 22)
- `--port, -p` : Listen port (default: 42388)

⚠️ When the service runs as root (LaunchDaemon, init.d), leaving `--username` empty yields **no username at all** — the account it would pick up is `root`, which is almost never the one people want to log in as, so it reports nothing rather than something misleading. **Always pass `--username` explicitly when deploying as a service.**

### Scan LAN Hosts

Scan every subnet this machine sits on:

```bash
udpscan client
```

Options:
- `--broadcast, -b` : Broadcast address (empty enumerates every interface)
- `--name, -n` : Keep only hosts whose nickname matches (nicknames may repeat, the count is reported)
- `--json` : Emit JSON for scripts to consume
- `--port, -p` : Port the **other side** listens on, probes go there (default: 42388)
- `--source-port` : Port **this machine** sends from and receives replies on (default 0 = random; pinning it lets the firewall open just one port, see "Deployment Notes")
- `--timeout, -t` : Scan timeout in seconds (default: 3)

**Exit code**: `0` when at least one host answered, `1` when none did, so it drops straight into a conditional:

```bash
if udpscan client --name xiaozhixiang > /dev/null; then echo "online"; else echo "offline"; fi
```

### Check Version and Protocol Identity

```bash
udpscan version
```

```
udpscan     v0.0.1
协议魔术字  UDPSCAN
默认端口    42388
Go 版本     go1.26.5
```

⭐ Whether two ends can talk depends on the **protocol magic string**, not the version number. Run this on both machines and compare: a mismatch is immediately visible, whereas in normal operation neither side reports anything at all (see "Deployment Notes" below).

## Example

```bash
# On Ubuntu (start detection service)
$ udpscan server --name xiaozhixiang --username yangyile
服务端启动 [xiaozhixiang] 监听端口 42388
对外通告的登录方式: yangyile@本机 (sshd 端口 22)

# On Mac (scan, covering every interface automatically)
$ udpscan client
扫描中... (广播: 172.16.91.255 10.42.0.255, 端口: 42388, 超时: 3s)
--------------------
发现: 172.16.91.98    xiaozhixiang
--------------------
共发现 1 台主机

ssh yangyile@172.16.91.98
```

That trailing ssh command is ready to paste — which is usually the very next thing anyone does after a scan.

For scripts, add `--json`:

```bash
$ udpscan client --json
{
  "broadcasts": ["172.16.91.255", "10.42.0.255"],
  "port": 42388,
  "timeout_seconds": 3,
  "started_at": "2026-08-08T02:42:18+08:00",
  "count": 1,
  "hosts": [
    {
      "ip": "172.16.91.98",
      "name": "xiaozhixiang",
      "username": "yangyile",
      "ssh_port": 22,
      "ssh_command": "ssh yangyile@172.16.91.98",
      "time": "2026-08-08T02:42:18+08:00"
    }
  ]
}
```

⭐ Those leading overview fields earn their keep **when nothing is found**: `broadcasts` records where the probes actually went, which separates "no host answered" from "the packets left through the wrong interface".

## How It Works

1. **Host Detection Service** listens on a UDP port and answers with its nickname, login username and sshd port
2. **Scan Command** enumerates the broadcast address of every local interface and probes each one, collecting responses from active hosts
3. Results show IP addresses, nicknames, and a ready-to-run ssh command

⭐ **The IP in that ssh command is decided by the scanner, not self-reported by the host.** A machine with several interfaces cannot know which of its addresses reached the scanner; only the source address the reply actually arrived from is guaranteed to be reachable.

## Deployment Notes

Each of the three issues below makes scanning fail **silently** — no error, no warning, just zero hosts found. Read this before deploying: it costs much less than packet-capturing after the fact.

### The firewall must open both directions

Opening only "destination port = scan port" is not enough. That admits the request packets (others scanning this host), but not the reply packets (this host scanning others).

- **Symptom**: others can detect this machine, yet this machine detects nobody
- **Cause**: the broadcast request goes to `x.x.x.255`, while replies arrive from concrete host addresses such as `x.x.x.190`. The source address does not match, so connection tracking cannot classify the reply as return traffic and the default policy drops it. The reply also lands on a random ephemeral port rather than the scan port, so the "destination port" rule cannot match it either
- **Fix**: add a second rule keyed on **source port = scan port**. Confine its destination to the local ephemeral port range (on Linux see `net.ipv4.ip_local_port_range`) so only the scanning process can receive, leaving other fixed UDP ports untouched

#### ⭐ Narrow that rule down to a single port with `--source-port`

The destination of the rule above spans the **entire ephemeral port range**:

```bash
# opens 32768-60999 — 28232 ports in total
-A INPUT -p udp --sport 42388 --dport 32768:60999 -j ACCEPT
```

The range has to be that wide because the client picks a **random port** for replies, unknowable ahead of time. And the rule keys on "source port = 42388" — **a value the sender picks freely**, so anyone can ride that opening to reach whatever else listens on those 28232 ports.

Pin the client's receiving port and the rule collapses to one port:

```bash
# scan with a fixed reply port
udpscan client --source-port 42389

# the firewall now needs exactly one port
-A INPUT -p udp --sport 42388 --dport 42389 -j ACCEPT
```

**28232 ports → 1 port.**

⚠️ The trade-off: two scans **cannot run at once** on the same machine (the port is taken). The command then **fails outright and suggests dropping the flag** — it deliberately does *not* fall back to a random port, because that would silently stop matching the firewall rule and turn into "sometimes it finds hosts, sometimes it doesn't", which is far harder to diagnose than an outright failure.

### Avoid the limited broadcast address, let enumeration handle it

`255.255.255.255` is the limited broadcast address, and the kernel picks the outgoing interface from the routing table. When a VPN, proxy or virtual adapter exists, the packet often leaves through that interface and never reaches the intended subnet — **and the send call still succeeds**.

**The default now sidesteps this entirely**: without `-b`, the client enumerates the directed broadcast address of every local interface (such as `192.168.1.255`) and probes each one, so every attached subnet gets covered. Point-to-point tunnels created by VPNs and proxies (`tun` / `utun`) carry no broadcast capability and are skipped automatically — the check reads the interface's broadcast capability flag rather than matching names, so whatever the tunnel calls itself makes no difference.

`-b` is now only needed to **restrict the scan to one subnet**, or to target a broadcast address **outside the local interfaces**. ⛔ Passing `255.255.255.255` by hand is discouraged: it invites the very problem described above back in.

### Both ends must run matching versions

The handshake depends on an agreed magic string, and the service replies only on an exact match. **Once that magic string changes, old and new builds go completely silent toward each other**: requests are still sent, the service still receives them, and mismatched payloads are dropped without a single log line on either side.

Upgrade every machine together. When diagnosing this class of problem, capturing one packet beats reading logs.

---

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 License

MIT License - see [LICENSE](LICENSE).

---

## 💬 Contact & Feedback

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Mistake reports?** Open an issue on GitHub with reproduction steps
- 💡 **Fresh ideas?** Create an issue to discuss
- 📖 **Documentation confusing?** Report it so we can improve
- 🚀 **Need new features?** Share the use cases to help us understand requirements
- ⚡ **Performance issue?** Help us optimize through reporting slow operations
- 🔧 **Configuration problem?** Ask questions about complex setups
- 📢 **Follow project progress?** Watch the repo to get new releases and features
- 🌟 **Success stories?** Share how this package improved the workflow
- 💬 **Feedback?** We welcome suggestions and comments

---

## 🔧 Development

New code contributions, follow this process:

1. **Fork**: Fork the repo on GitHub (using the webpage UI).
2. **Clone**: Clone the forked project (`git clone https://github.com/yourname/repo-name.git`).
3. **Navigate**: Navigate to the cloned project (`cd repo-name`)
4. **Branch**: Create a feature branch (`git checkout -b feature/xxx`).
5. **Code**: Implement the changes with comprehensive tests
6. **Testing**: (Golang project) Ensure tests pass (`go test ./...`) and follow Go code style conventions
7. **Documentation**: Update documentation to support client-facing changes
8. **Stage**: Stage changes (`git add .`)
9. **Commit**: Commit changes (`git commit -m "Add feature xxx"`) ensuring backward compatible code
10. **Push**: Push to the branch (`git push origin feature/xxx`).
11. **PR**: Open a merge request on GitHub (on the GitHub webpage) with detailed description.

Please ensure tests pass and include relevant documentation updates.

---

## 🌟 Support

Welcome to contribute to this project via submitting merge requests and reporting issues.

**Project Support:**

- ⭐ **Give GitHub stars** if this project helps you
- 🤝 **Share with teammates** and (golang) programming friends
- 📝 **Write tech blogs** about development tools and workflows - we provide content writing support
- 🌟 **Join the ecosystem** - committed to supporting open source and the (golang) development scene

**Have Fun Coding with this package!** 🎉🎉🎉

<!-- TEMPLATE (EN) CLOSE: STANDARD PROJECT FOOTER -->

---

## GitHub Stars

[![Stargazers](https://starchart.cc/yylego/udpscan.svg?variant=adaptive)](https://starchart.cc/yylego/udpscan)
