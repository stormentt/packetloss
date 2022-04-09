package wrappers

import (
	"crypto/subtle"
	"fmt"

	proto "github.com/stormentt/packetloss/packet"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/openpgp/packet"
)

type CryptoError struct {
	Reason string
	Err    error
}

func (c *CryptoError) Error() string {
	if c.Err != nil {
		return fmt.Sprintf("crypto error: %s\n", c.Reason)
	} else {
		return fmt.Sprintf("crypto error: %s (%v)\n", c.Reason, c.Err)
	}
}

func EncodePacket(p packet.Packet, hkey []byte) ([]byte, error) {
	data, err := proto.Marshal(p)
	if err != nil {
		return nil, err
	}

	calc_hmac := make([]byte, 64)
	err := hmac_data(calc_hmac, hkey, data)

	out := make([]byte, 64+len(data))
	copy(out, calc_hmac)
	copy(out[64:], data)

	return out, nil
}

func DecodePacket(data []byte, hkey []byte, p *packet.Packet) error {
	data_hmac := make([]byte, 64)
	copy(data_hmac, data[:64])

	calc_hmac := make([]byte, 64)
	err := hmac_data(calc_hmac, hkey, data[64:])
	if err != nil {
		return &CryptoError{
			Reason: "could not calculate hmac",
			Err:    err,
		}
	}

	if subtle.ConstantTimeCompare(calc_hmac, data_hmac) != 1 {
		return &CryptoError{
			Reason: "calculated hash and expected hash did not match",
			Err:    nil,
		}
	}

	err = proto.Unmarshal(data[64:], p)
	if err != nil {
		return err
	}

	return nil
}

func hmac_data(out, hkey, data []byte) error {
	hasher, err := blake2b.New512(hkey)
	if err != nil {
		return err
	}

	hasher.Write(data)
	copy(out, hasher.Sum(nil))
	return nil
}
