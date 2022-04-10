package client

// ClientRecord is a collection of stats for clients
type ClientRecord struct {
	Sent  uint64
	Acked uint64

	LastSent uint64
	LastAck  uint64
}

// NewClientRecord creates a new ClientRecord object
func NewClientRecord() *ClientRecord {
	return &ClientRecord{}
}

// Send records that the serial number was sent
func (cr *ClientRecord) Send(serial uint64) {
	cr.Sent++
	cr.LastSent = serial
}

// Ack records that the serial number was acknowledged
func (cr *ClientRecord) Ack(serial uint64) {
	cr.Acked++
	cr.LastAck = serial
}

// Remediate returns sums for various stats
func (cr *ClientRecord) Remediate() *ClientStats {
	return &ClientStats{
		Sent:          cr.Sent,
		Acked:         cr.Acked,
		Missed:        cr.Sent - cr.Acked,
		MissedPercent: float64(cr.Sent-cr.Acked) / float64(cr.Sent) * 100.0,
	}
}

// Reset resets the Sent and Ack counters
func (cr *ClientRecord) Reset() {
	cr.Sent = 0
	cr.Acked = 0
}

type ClientStats struct {
	Sent          uint64
	Acked         uint64
	Missed        uint64
	MissedPercent float64
}
