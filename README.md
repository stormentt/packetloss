# Overview
Packetloss is a tool for measuring... well, packet loss. It operates in two modes, client and server.

## Client Mode
In client mode, packetloss continuously sends UDP packets with an incrementing numeric ID and expects acknowledgements for each packet that it sends. Packetloss will keep a count of sent packets and received acknowledgements and periodically print statistics showing how many packets were sent, how many acked, and how many acks were expected but never received.

## Server Mode
In server mode, packetloss listens for UDP packets, and upon receiving (valid) packets it sends back a response acknowledging that it received that packet. If it received a packet with a serial number that is higher than expected, it infers that it has missed packets and will record that.

# Configuration
Packetloss requires either a `packetloss.yml` or config parameters to be passed via command line.

**packetloss.yml**
```yaml
key:    "A RANDOM KEY"    #  used for HMAC
local:  ":6666"           #  local address to listen on
remote: "localhost:6666"  #  remote address to send packets to
```

# Usage
`packetloss client` for client mode

`packetloss server` for server mode
