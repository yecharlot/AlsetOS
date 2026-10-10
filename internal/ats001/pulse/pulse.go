package pulse

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/yecharlot/AlsetOS/internal/ats001/identity"
)

type Message struct {
	SourceOID      string                 `json:"SourceOID"`
	DestinationOID string                 `json:"DestinationOID"`
	Timestamp      int64                  `json:"Timestamp"`
	Nonce          string                 `json:"Nonce"`
	Signature      string                 `json:"Signature"`
	Capabilities   []string               `json:"Capabilities"`
	PayloadCID     string                 `json:"PayloadCID"`
	Payload        map[string]interface{} `json:"Payload"`
}

func (m *Message) Canonical() []byte {
	cp := *m
	cp.Signature = ""
	b, _ := json.Marshal(cp)
	return b
}

func Make(src identity.RootIdentity, dest string, payload map[string]interface{}, caps []string) Message {
	nb := make([]byte, 8)
	_, _ = rand.Read(nb)
	m := Message{
		SourceOID:      src.OID,
		DestinationOID: dest,
		Timestamp:      time.Now().Unix(),
		Nonce:          hex.EncodeToString(nb),
		Capabilities:   caps,
		PayloadCID:     identity.CID(payload),
		Payload:        payload,
	}
	m.Signature = src.Keys.Sign(m.Canonical())
	return m
}

func Verify(src identity.RootIdentity, m Message) bool {
	return src.Keys.Verify(m.Canonical(), m.Signature)
}
