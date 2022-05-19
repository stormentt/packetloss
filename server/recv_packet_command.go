package server

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type RecvPacketCommand struct {
	ws wrapSerial

	oldStats *ServerStats
}

func newRecvPacketCommand(ws wrapSerial) *RecvPacketCommand {
	oldStats := &ServerStats{}

	return &RecvPacketCommand{
		ws,
		oldStats,
	}
}

func (cmd *RecvPacketCommand) Do(sm *StatsMap) error {
	stats := sm.Get(cmd.ws.ClientID)
	stats.Clone(cmd.oldStats)

	if stats.LastSerial >= cmd.ws.Serial {
		log.WithFields(log.Fields{
			"LastSerial": stats.LastSerial,
			"Serial":     cmd.ws.Serial,
		}).Warn("packet received with old serial")

		return RecvOldSerialErr{
			LastSerial: stats.LastSerial,
			Serial:     cmd.ws.Serial,
		}
	}

	dSerial := cmd.ws.Serial - stats.LastSerial
	if dSerial > 1 {
		log.WithFields(log.Fields{
			"LastSerial": stats.LastSerial,
			"Serial":     cmd.ws.Serial,
			"dSerial":    dSerial,
		}).Info("missed packets")

		stats.Missed += dSerial
	}

	stats.Received++
	stats.LastSerial = cmd.ws.Serial
	stats.LastUpdated = time.Now()

	return nil
}

func (cmd *RecvPacketCommand) Undo(sm *StatsMap) error {
	stats := sm.Get(cmd.ws.ClientID)
	cmd.oldStats.Clone(stats)

	return nil
}
