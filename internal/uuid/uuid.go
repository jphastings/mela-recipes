package uuid

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type UUID [16]byte

// NewUUID creates a new UUIDv4. If Empty string is provided then random bits are used; if a string is provided then the (cryptographically random & deterministic) bits of the SHA256 of the given string are used instead.
func NewUUID(in string) (UUID, error) {
	var u UUID
	if in == "" {
		r := make([]byte, 16)
		n, err := rand.Read(r)
		if err != nil {
			return u, err
		}
		if n != 16 {
			if err != nil {
				return u, fmt.Errorf("not enough random data available to generate UUID")
			}
		}
		copy(u[:], r[:])
	} else {
		h := sha256.Sum256([]byte(in))
		copy(u[:], h[:16])
	}

	// Set version 4 UUID (pure random)
	u[6] &= 0x0f
	u[6] |= 0x40
	// Set variant (RFC 4122)
	u[8] &= 0x3f
	u[8] |= 0x80
	return u, nil
}

func (uuid UUID) String() string {
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%12x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:16],
	)
}

func (uuid UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uuid.String())
}
