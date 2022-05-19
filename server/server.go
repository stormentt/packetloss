package server

import (
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	packet "github.com/stormentt/packetloss/packet"
	wrapper "github.com/stormentt/packetloss/wrapper"
)

type wrapSerial struct {
	Serial   uint64
	From     *net.UDPAddr
	ClientID string
}

// Listen listens for new packets and acks them, as well as keeping records
// hkey is used to validate incoming messages via message authentication codes
func Listen(hkey []byte, laddr *net.UDPAddr) error {
	log.WithFields(log.Fields{
		"UpdateTime": viper.GetDuration("update-time"),
		"CullTime":   viper.GetDuration("cull_time"),
	}).Debug("server params")
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}

	log.Info("handling connections")
	defer conn.Close()

	sMap := NewStatsMap()
	lastCull := time.Now()
	lastStatsPrint := time.Now()

	ch := make(chan StatsCommand, 10)
	go handleRecv(conn, hkey, ch)

	for cmd := range ch {
		err = cmd.Do(sMap)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Error("unable to execute command")

			continue
		}

		if time.Since(lastCull) > viper.GetDuration("cull_time") {
			sMap.Cull()
			lastCull = time.Now()
		}

		if time.Since(lastStatsPrint) > viper.GetDuration("update-time") {
			sMap.Print()
			lastStatsPrint = time.Now()
		}
	}

	return nil
}

// handleRecv receives packets from conn and acknowledges them
// received serial numbers are sent over ch for record keeping
func handleRecv(conn *net.UDPConn, hkey []byte, ch chan<- StatsCommand) {
	for {
		buff := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buff)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("could not read from UDP conn")
			continue
		}

		log.WithFields(log.Fields{
			"n":    n,
			"addr": addr,
		}).Trace("received packet")

		p := &packet.Packet{}
		err = wrapper.DecodePacket(buff, n, hkey, p)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("could not decode packet")

			continue
		}

		log.WithFields(log.Fields{
			"PacketType": p.PacketType.String(),
			"Serial":     p.Serial,
			"ClientID":   p.ClientID,
		}).Trace("decoded packet")

		switch p.PacketType {
		case packet.PacketType_REQPACKET:
			ws := wrapSerial{
				Serial:   p.Serial,
				From:     addr,
				ClientID: p.ClientID,
			}

			reqCmd := newRecvPacketCommand(ws)
			ch <- reqCmd

			sendAck(conn, hkey, ch, ws)
		case packet.PacketType_RESETPACKET:
			ws := wrapSerial{
				Serial:   p.Serial,
				From:     addr,
				ClientID: p.ClientID,
			}

			resetCmd := newResetPacketCommand(ws)
			ch <- resetCmd
		default:
			log.WithFields(log.Fields{
				"type": p.PacketType.String(),
			}).Warn("received an unexpected packet type")

			continue
		}
	}
}

func sendAck(conn *net.UDPConn, hkey []byte, ch chan<- StatsCommand, ws wrapSerial) {
	ackPacket := packet.Packet{
		Serial:     ws.Serial,
		PacketType: packet.PacketType_ACKPACKET,
		ClientID:   ws.ClientID,
	}

	data, err := wrapper.EncodePacket(&ackPacket, hkey)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("could not encode ack packet")

		return
	}

	_, err = conn.WriteTo(data, ws.From)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("could not send ack packet")

		return
	}

	ackCmd := newAckPacketCommand(ws)

	ch <- ackCmd
}
