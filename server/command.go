package server

type StatsCommand interface {
	Do(sm *StatsMap) error
	Undo(sm *StatsMap) error
}
