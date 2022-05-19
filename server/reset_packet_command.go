package server

type ResetPacketCommand struct {
	ws wrapSerial

	oldStats *ServerStats
}

func newResetPacketCommand(ws wrapSerial) *ResetPacketCommand {
	oldStats := &ServerStats{}

	return &ResetPacketCommand{
		ws,
		oldStats,
	}
}

func (cmd *ResetPacketCommand) Do(sm *StatsMap) error {
	stats := sm.Get(cmd.ws.ClientID)
	stats.Clone(cmd.oldStats)

	stats.Reset()

	return nil
}

func (cmd *ResetPacketCommand) Undo(sm *StatsMap) error {
	stats := sm.Get(cmd.ws.ClientID)
	cmd.oldStats.Clone(stats)

	return nil
}
