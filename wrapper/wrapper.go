package wrappers

import (
	"crypto/subtle"
	"fmt"

	log "github.com/sirupsen/logrus"
	packet "github.com/stormentt/packetloss/packet"
	"golang.org/x/crypto/blake2b"
	"google.golang.org/protobuf/proto"
)

type CryptoError struct {
	Reason string
	Err    error
}

func (c *CryptoError) Error() string {
	if c.Err == nil {
		return fmt.Sprintf("crypto error: %s", c.Reason)
	} else {
		return fmt.Sprintf("crypto error: %s (%v)", c.Reason, c.Err)
	}
}

// EncodePacket takes a protobuf packet and prepends a message authentication code to it
// hkey is used to create a keyed Blake2b hash
func EncodePacket(p *packet.Packet, hkey []byte) ([]byte, error) {
	data, err := proto.Marshal(p)
	if err != nil {
		return nil, err
	}

	calc_hmac := make([]byte, 64)
	err = hmac_data(calc_hmac, hkey, data)

	log.WithFields(log.Fields{
		"HMAC": fmt.Sprintf("%X", calc_hmac),
	}).Debug("Calculated HMAC for Packet")

	out := make([]byte, 64+len(data))
	copy(out, calc_hmac)
	copy(out[64:], data)

	return out, nil
}

// DecodePacket takes a blob of data, validates it, and decodes it into a protobuf packet.
// hkey is used to create a keyed Blake2b hash
// packets are rejected if their message authentication code is invalid
func DecodePacket(data []byte, n int, hkey []byte, p *packet.Packet) error {
	data_hmac := make([]byte, 64)
	copy(data_hmac, data[:64])

	calc_hmac := make([]byte, 64)
	err := hmac_data(calc_hmac, hkey, data[64:n])
	if err != nil {
		return &CryptoError{
			Reason: "could not calculate hmac",
			Err:    err,
		}
	}

	if subtle.ConstantTimeCompare(calc_hmac, data_hmac) != 1 {
		log.WithFields(log.Fields{
			"Calculated": fmt.Sprintf("%X", calc_hmac),
			"Expected":   fmt.Sprintf("%X", data_hmac),
		}).Debug("hash mismatch")

		return &CryptoError{
			Reason: "calculated hash and expected hash did not match",
			Err:    nil,
		}
	}

	err = proto.Unmarshal(data[64:n], p)
	if err != nil {
		return err
	}

	return nil
}

// hmac_data calculates the message authentication code for a blob of data, using hkey as a key
func hmac_data(out, hkey, data []byte) error {
	hasher, err := blake2b.New512(hkey)
	if err != nil {
		return err
	}

	hasher.Write(data)
	copy(out, hasher.Sum(nil))
	return nil
}
