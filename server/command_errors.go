package server

import "fmt"

type RecvOldSerialErr struct {
	LastSerial uint64
	Serial     uint64
}

func (e RecvOldSerialErr) Error() string {
	return fmt.Sprintf("packet received with old serial: %d >= %d", e.LastSerial, e.Serial)
}
