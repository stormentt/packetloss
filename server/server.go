package server

import (
	"net"
	"time"

	log "github.com/sirupsen/logrus"
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
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}

	log.Info("handling connections")
	defer conn.Close()

	sMap := NewStatsMap()
	lastCull := time.Now()
	lastStatsPrint := time.Now()

	ch := make(chan wrapSerial, 10)
	go handleRecv(conn, hkey, ch)

	for ws := range ch {
		stats := sMap.Get(ws.ClientID)

		if stats.LastSerial >= ws.Serial {
			if ws.Serial != 0 {
				log.WithFields(log.Fields{
					"LastSerial": stats.LastSerial,
					"Serial":     ws.Serial,
				}).Warn("packet received with old serial")
			}

			continue
		}

		dSerial := ws.Serial - stats.LastSerial
		if dSerial > 1 {
			log.WithFields(log.Fields{
				"LastSerial": stats.LastSerial,
				"Serial":     ws.Serial,
				"dSerial":    dSerial,
			}).Debug("recording missed packets")

			stats.Missed += dSerial
		}

		stats.Received++
		stats.LastSerial = ws.Serial
		stats.LastUpdated = time.Now()

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
		}

		_, err = conn.WriteTo(data, ws.From)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
			}).Error("could not send ack packet")
		}

		stats.LastAck = ws.Serial

		if time.Since(lastCull) > time.Minute*10 {
			sMap.Cull()
			lastCull = time.Now()
		}

		if time.Since(lastStatsPrint) > time.Minute*10 {
			sMap.Print()
			lastStatsPrint = time.Now()
		}
	}

	return nil
}

// handleRecv receives packets from conn and acknowledges them
// received serial numbers are sent over ch for record keeping
func handleRecv(conn *net.UDPConn, hkey []byte, ch chan<- wrapSerial) {
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

		if p.PacketType != packet.PacketType_REQPACKET {
			log.WithFields(log.Fields{
				"type": p.PacketType.String(),
			}).Warn("received a non-REQPACKET type packet")
			continue
		}

		ws := wrapSerial{
			Serial:   p.Serial,
			From:     addr,
			ClientID: p.ClientID,
		}

		ch <- ws
	}
}
