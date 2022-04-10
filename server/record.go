package server

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type StatsMap struct {
	internal map[string]*ServerStats
}

func NewStatsMap() *StatsMap {
	return &StatsMap{
		internal: make(map[string]*ServerStats),
	}
}

func (sm *StatsMap) Get(client string) *ServerStats {
	if stats, ok := sm.internal[client]; ok {
		stats.LastUpdated = time.Now()
		return stats
	} else {
		sm.internal[client] = &ServerStats{}
		return sm.internal[client]
	}
}

func (sm *StatsMap) Print() {
	for client, stats := range sm.internal {
		totalPackets := stats.Received + stats.Missed
		percentLoss := float64(stats.Missed) / float64(totalPackets) * 100.0

		log.WithFields(log.Fields{
			"Total":       totalPackets,
			"Missed":      stats.Missed,
			"PercentMiss": fmt.Sprintf("%0.2f", percentLoss),
			"ClientID":    client,
			"LastUpdate":  stats.LastUpdated,
			"Timestamp":   time.Now(),
		}).Info("stats")
	}
}

func (sm *StatsMap) Cull() {
	for id, stats := range sm.internal {
		if stats.Cullable() {
			delete(sm.internal, id)
		}
	}
}

type ServerStats struct {
	Received uint64
	Missed   uint64

	LastSerial uint64
	LastAck    uint64

	LastUpdated time.Time
}

func (stats *ServerStats) Reset() {
	stats.Received = 0
	stats.Missed = 0

	stats.LastSerial = 0
	stats.LastAck = 0

	stats.LastUpdated = time.Now()
}

func (stats *ServerStats) Cullable() bool {
	if time.Since(stats.LastUpdated) > time.Minute*30 {
		return true
	}

	return false
}
