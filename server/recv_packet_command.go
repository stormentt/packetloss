package server

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type RecvPacketCommand struct {
	ws wrapSerial

	oldStats *ServerStats
}

func newRecvPacketCommand(ws wrapSerial) RecvPacketCommand {
	oldStats := &ServerStats{}

	return RecvPacketCommand{
		ws,
		oldStats,
	}
}

func (rpc *RecvPacketCommand) Do(sm *StatsMap) error {
	stats := sm.Get(rpc.ws.ClientID)
	stats.Clone(rpc.oldStats)

	if stats.LastSerial >= rpc.ws.Serial {
		log.WithFields(log.Fields{
			"LastSerial": stats.LastSerial,
			"Serial":     rpc.ws.Serial,
		}).Warn("packet received with old serial")

		return RecvOldSerialErr{
			LastSerial: stats.LastSerial,
			Serial:     rpc.ws.Serial,
		}
	}

	dSerial := rpc.ws.Serial - stats.LastSerial
	if dSerial > 1 {
		log.WithFields(log.Fields{
			"LastSerial": stats.LastSerial,
			"Serial":     rpc.ws.Serial,
			"dSerial":    dSerial,
		}).Info("missed packets")

		stats.Missed += dSerial
	}

	stats.Received++
	stats.LastSerial = rpc.ws.Serial
	stats.LastUpdated = time.Now()

	return nil
}

func (rpc *RecvPacketCommand) Undo(sm *StatsMap) error {
	stats := sm.Get(rpc.ws.ClientID)
	rpc.oldStats.Clone(stats)

	return nil
}
