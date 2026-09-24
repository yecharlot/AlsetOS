package p2p

import (
	"fmt"

	"github.com/ipfs/go-cid"
	multihash "github.com/multiformats/go-multihash"
)

func CIDRoot(rootCID string) (cid.Cid, error) {
	if rootCID == "" {
		return cid.Cid{}, fmt.Errorf("RootCID vacío")
	}
	digest, err := multihash.Sum([]byte(rootCID), multihash.SHA2_256, -1)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("crear multihash RootCID: %w", err)
	}
	return cid.NewCidV1(cid.Raw, digest), nil
}
