package p2p

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ipfs/go-cid"
	multihash "github.com/multiformats/go-multihash"
)

func CIDRoot(rootCID string) (cid.Cid, error) {
	rootCID = strings.TrimSpace(rootCID)
	const prefijo = "rootcid:"
	if !strings.HasPrefix(rootCID, prefijo) {
		return cid.Cid{}, fmt.Errorf("RootCID inválido: %q", rootCID)
	}

	digest, err := hex.DecodeString(strings.TrimPrefix(rootCID, prefijo))
	if err != nil || len(digest) != 32 {
		return cid.Cid{}, fmt.Errorf("digest RootCID inválido")
	}

	contenido, err := multihash.Encode(digest, multihash.SHA2_256)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("crear multihash RootCID: %w", err)
	}
	return cid.NewCidV1(cid.Raw, contenido), nil
}
