package server

type AckPacketCommand struct {
	ws wrapSerial

	oldStats *ServerStats
}

func newAckPacketCommand(ws wrapSerial) *AckPacketCommand {
	oldStats := &ServerStats{}

	return &AckPacketCommand{
		ws,
		oldStats,
	}
}

func (cmd *AckPacketCommand) Do(sm *StatsMap) error {
	stats := sm.Get(cmd.ws.ClientID)
	stats.Clone(cmd.oldStats)

	stats.LastAck = cmd.ws.Serial

	return nil
}

func (cmd *AckPacketCommand) Undo(sm *StatsMap) error {
	stats := sm.Get(cmd.ws.ClientID)
	cmd.oldStats.Clone(stats)

	return nil
}
