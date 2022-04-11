package client

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type PacketRecord struct {
	Serial uint64

	Sent     bool
	SentTime time.Time

	Acked     bool
	AckedTime time.Time
}

// ClientRecord is a collection of stats for clients
type ClientRecord struct {
	Packets map[uint64]*PacketRecord

	LastSent uint64
	LastAck  uint64
}

// NewClientRecord creates a new ClientRecord object
func NewClientRecord() *ClientRecord {
	return &ClientRecord{
		Packets: make(map[uint64]*PacketRecord),
	}
}

// Send records that the serial number was sent
func (cr *ClientRecord) Send(serial uint64, ts time.Time) {
	log.WithFields(log.Fields{
		"Serial":    serial,
		"Timestamp": ts,
	}).Debug("send")

	cr.Packets[serial] = &PacketRecord{
		Serial:    serial,
		Sent:      true,
		SentTime:  ts,
		Acked:     false,
		AckedTime: time.Time{},
	}

	cr.LastSent = serial
}

// Ack records that the serial number was acknowledged
func (cr *ClientRecord) Ack(serial uint64, ts time.Time) {
	log.WithFields(log.Fields{
		"Serial":    serial,
		"Timestamp": ts,
	}).Debug("ack")

	if pr, ok := cr.Packets[serial]; ok {
		pr.Acked = true
		pr.AckedTime = ts
	} else {
		cr.Packets[serial] = &PacketRecord{
			Serial:    serial,
			Sent:      false,
			SentTime:  time.Time{},
			Acked:     true,
			AckedTime: ts,
		}
	}

	cr.LastAck = serial
}

// Remediate returns sums for various stats
func (cr *ClientRecord) Remediate() *ClientStats {
	var Total uint64

	var TotalSent uint64
	var TotalAcked uint64

	var SentAndAcked uint64
	var SentNotAcked uint64
	var AckedNotSent uint64

	var TotalRTT time.Duration
	var MinRTT time.Duration
	var MaxRTT time.Duration

	for _, pr := range cr.Packets {
		if !pr.Sent && !pr.Acked {
			// should never happen
			log.WithFields(log.Fields{
				"Serial": pr.Serial,
			}).Warn("recorded packet never sent and never acked. this should never happen")

			continue
		}

		Total++

		if pr.Sent {
			TotalSent++
		}

		if pr.Acked {
			TotalAcked++
		}

		if pr.Sent && pr.Acked {
			SentAndAcked++

			RTT := pr.AckedTime.Sub(pr.SentTime)
			TotalRTT += RTT

			if RTT < MinRTT || MinRTT == time.Duration(0) {
				MinRTT = RTT
			}

			if RTT > MaxRTT || MaxRTT == time.Duration(0) {
				MaxRTT = RTT
			}
		}

		if pr.Sent && !pr.Acked {
			SentNotAcked++
		}

		if !pr.Sent && pr.Acked {
			AckedNotSent++
		}
	}

	var AvgRTT time.Duration
	if SentAndAcked != 0 {
		AvgRTT = TotalRTT / time.Duration(SentAndAcked)
	}

	SAAPercent := float64(SentAndAcked) / float64(Total) * 100.0
	SNAPercent := float64(SentNotAcked) / float64(Total) * 100.0
	ANSPercent := float64(AckedNotSent) / float64(Total) * 100.0

	return &ClientStats{
		Total,
		TotalSent,
		TotalAcked,

		SentAndAcked,
		SAAPercent,

		SentNotAcked,
		SNAPercent,

		AckedNotSent,
		ANSPercent,

		AvgRTT,
		MinRTT,
		MaxRTT,
	}
}

// Reset resets the Sent and Ack counters
func (cr *ClientRecord) Reset() {
	cr.Packets = make(map[uint64]*PacketRecord)
}

type ClientStats struct {
	Total uint64

	TotalSent  uint64
	TotalAcked uint64

	SentAndAcked uint64
	SAAPercent   float64

	SentNotAcked uint64
	SNAPercent   float64

	AckedNotSent uint64
	ANSPercent   float64

	AvgRTT time.Duration
	MinRTT time.Duration
	MaxRTT time.Duration
}
