# goping
Goping is a lightweight CLI tool designed to replicate the basic features of the ping cli. The tool was designed to have minimal dependencies, only using [Cobra](https://github.com/spf13/cobra) for CLI utilities.

Monitoring incoming ICMP requests requires root access and so the compiled binary resulting from this codebase has to be run with sudo. (ping has setuid)

## How To Use
```bash
sudo goping <hostname>
sudo goping <ip>
```

## Implemented Features
- IPv4 ICMP Echo packets are sent out in an infinite loop.
- Packet losses and RTT times are reported for each message.
- On ctrl-C, goping displays statistics for the echo requests.
