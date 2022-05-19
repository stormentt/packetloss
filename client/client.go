package client

import (
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	packet "github.com/stormentt/packetloss/packet"
	wrapper "github.com/stormentt/packetloss/wrapper"
)

// wrapSerial
type wrapSerial struct {
	Serial uint64
	Type   packet.PacketType

	Timestamp time.Time
}

// Start sends UDP packets to raddr and keeps track of sent packets & acknowledgements.
// hkey is used to create message authentication codes
func Start(hkey []byte, raddr *net.UDPAddr) error {
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
			cr.Send(ws.Serial, ws.Timestamp)
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

			cr.Ack(ws.Serial, ws.Timestamp)

		default:
			log.WithFields(log.Fields{
				"Type": ws.Type,
			}).Warn("receieved an unexpected packet type")

			continue
		}

		if time.Since(lastRemediation) > viper.GetDuration("update-time") {
			stats := cr.Remediate()
			cr.Reset()

			log.WithFields(log.Fields{
				"Total":        stats.Total,
				"Sent":         stats.TotalSent,
				"Acked":        stats.TotalAcked,
				"SentAndAcked": stats.SentAndAcked,
				"SentNotAcked": stats.SentNotAcked,
				"AckedNotSent": stats.AckedNotSent,
			}).Info("Totals")

			log.WithFields(log.Fields{
				"SentAndAcked": fmt.Sprintf("%.2f", stats.SAAPercent),
				"SentNotAcked": fmt.Sprintf("%.2f", stats.SNAPercent),
				"AckedNotSent": fmt.Sprintf("%.2f", stats.ANSPercent),
			}).Info("Percents")

			log.WithFields(log.Fields{
				"Avg": stats.AvgRTT,
				"Min": stats.MinRTT,
				"Max": stats.MaxRTT,
			}).Info("RTT")

			lastRemediation = time.Now()
		}
	}

	return nil
}

// sendPackets sends packets to conn
// hkey is used to create message authentication codes for these packets
// sent packets have their serial numbers sent over ch, to be used for recordkeeping
func sendPackets(conn *net.UDPConn, hkey []byte, ch chan<- wrapSerial) {
	clientID := viper.GetString("client_id")
	if len(clientID) == 0 {
		log.Debug("no client ID specified, generating random")
		clientID = uuid.New().String()
	}

	if len(clientID) > 64 {
		log.WithFields(log.Fields{
			"ClientID": clientID,
			"Length":   len(clientID),
		}).Fatal("clientID length too long, max length 64")

		return
	}

	var serial uint64 = 1

	resetP := &packet.Packet{
		PacketType: packet.PacketType_RESETPACKET,
		Serial:     serial,
		ClientID:   clientID,
	}

	err := sendPacket(conn, hkey, resetP)
	if err != nil {
		return
	}

	for {
		time.Sleep(viper.GetDuration("packet_time"))

		p := &packet.Packet{
			PacketType: packet.PacketType_REQPACKET,
			Serial:     serial,
			ClientID:   clientID,
		}

		err = sendPacket(conn, hkey, p)
		if err != nil {
			continue
		}

		ws := wrapSerial{
			Serial:    serial,
			Type:      packet.PacketType_REQPACKET,
			Timestamp: time.Now(),
		}

		ch <- ws

		serial++

	}
}

func sendPacket(conn *net.UDPConn, hkey []byte, p *packet.Packet) error {
	data, err := wrapper.EncodePacket(p, hkey)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("unable to encode packet")

		return err
	}

	_, err = conn.Write(data)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("unable to send packet")

		return err
	}

	return nil
}

// recvPackets receives UDP packets from conn
// hkey is used to validate incoming message authentication codes
// received acknowledgements have their serial numbers sent over ch to be used for recordkeeping
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

		ts := time.Now()

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
			Serial:    p.Serial,
			Type:      p.PacketType,
			Timestamp: ts,
		}

		ch <- ws
	}
}
