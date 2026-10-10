package identity

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// ATS-001: OID + RootCID + claves. Referencia HMAC-SHA256 (stdlib).
type KeyPair struct {
	Public  string `json:"public"`
	Private string `json:"-"`
}

func GenerateKeys() KeyPair {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	priv := hex.EncodeToString(b)
	h := sha256.Sum256([]byte(priv))
	return KeyPair{Public: "ed25519-ref:" + hex.EncodeToString(h[:]), Private: priv}
}

func (k KeyPair) Sign(msg []byte) string {
	m := hmac.New(sha256.New, []byte(k.Private))
	m.Write(msg)
	return "hmac-sha256:" + hex.EncodeToString(m.Sum(nil))
}

func (k KeyPair) Verify(msg []byte, sig string) bool {
	if len(sig) < 12 || sig[:12] != "hmac-sha256:" {
		return false
	}
	m := hmac.New(sha256.New, []byte(k.Private))
	m.Write(msg)
	return hmac.Equal([]byte(sig[12:]), []byte(hex.EncodeToString(m.Sum(nil))))
}

func CID(v interface{}) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return "bafy" + hex.EncodeToString(h[:22])
}

type RootIdentity struct {
	OID       string  `json:"oid"`
	RootCID   string  `json:"root_cid"`
	Keys      KeyPair `json:"-"`
	CreatedAt int64   `json:"created_at"`
}

func Create(name string, desc map[string]interface{}) RootIdentity {
	keys := GenerateKeys()
	root := CID(map[string]interface{}{"name": name, "desc": desc, "kpub": keys.Public})
	oid := fmt.Sprintf("alset:org:%s:%s", name, root[4:12])
	return RootIdentity{OID: oid, RootCID: root, Keys: keys, CreatedAt: time.Now().Unix()}
}
