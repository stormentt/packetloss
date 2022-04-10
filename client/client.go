package client

import (
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	packet "github.com/stormentt/packetloss/packet"
	wrapper "github.com/stormentt/packetloss/wrapper"
)

type wrapSerial struct {
	Serial uint64
	Type   packet.PacketType
}

func Send(hkey []byte, raddr *net.UDPAddr) error {
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil
	}

	defer conn.Close()

	cr := NewClientRecord()
	lastRemediation := time.Now()

	ch := make(chan wrapSerial, 10)

	go recvPackets(conn, hkey, ch)
	go sendPackets(conn, hkey, ch)

	for ws := range ch {
		switch ws.Type {
		case packet.PacketType_REQPACKET:
			cr.Send(ws.Serial)
		case packet.PacketType_ACKPACKET:
			if ws.Serial > cr.LastSent {
				log.WithFields(log.Fields{
					"Serial":   ws.Serial,
					"LastSent": cr.LastSent,
				}).Warn("received an ack for a packet we haven't sent yet")
				continue
			}

			if ws.Serial <= cr.LastAck {
				log.WithFields(log.Fields{
					"Serial":  ws.Serial,
					"LastAck": cr.LastAck,
				}).Warn("received an ack for an old packet")
				continue
			}

			cr.Ack(ws.Serial)

		default:
			log.WithFields(log.Fields{
				"Type": ws.Type,
			}).Warn("receieved an unexpected packet type")

			continue
		}

		if time.Since(lastRemediation) > time.Minute*1 {
			stats := cr.Remediate()
			cr.Reset()
			log.WithFields(log.Fields{
				"Sent":          stats.Sent,
				"Acked":         stats.Acked,
				"Missed":        stats.Missed,
				"MissedPercent": fmt.Sprintf("%0.2f", stats.MissedPercent),
				"Timestamp":     time.Now(),
			}).Info("stats")

			lastRemediation = time.Now()
		}
	}

	return nil
}

func sendPackets(conn *net.UDPConn, hkey []byte, ch chan<- wrapSerial) {
	clientID := uuid.New()
	var serial uint64 = 0

	for {
		p := &packet.Packet{
			PacketType: packet.PacketType_REQPACKET,
			Serial:     serial,
			ClientID:   clientID.String(),
		}

		data, err := wrapper.EncodePacket(p, hkey)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Fatal("unable to encode packet")
		}

		_, err = conn.Write(data)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Fatal("unable to send packet")
		}

		ws := wrapSerial{
			Serial: serial,
			Type:   packet.PacketType_REQPACKET,
		}

		ch <- ws

		serial++

		time.Sleep(500 * time.Millisecond)
	}
}

func recvPackets(conn *net.UDPConn, hkey []byte, ch chan<- wrapSerial) {
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
		}).Debug("received packet")

		p := &packet.Packet{}
		err = wrapper.DecodePacket(buff, n, hkey, p)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("could not decode packet")
			continue
		}

		if p.PacketType != packet.PacketType_ACKPACKET {
			log.WithFields(log.Fields{
				"type": p.PacketType.String(),
			}).Warn("received a non-ACKPACKET type packet")
			continue
		}

		ws := wrapSerial{
			Serial: p.Serial,
			Type:   p.PacketType,
		}

		ch <- ws
	}
}
