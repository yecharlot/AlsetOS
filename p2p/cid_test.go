package p2p

import (
	"strings"
	"testing"
)

func TestCIDRootRepresentaRootCID(t *testing.T) {
	root := "rootcid:" + strings.Repeat("ab", 32)
	c, err := CIDRoot(root)
	if err != nil { t.Fatal(err) }
	if !c.Defined() { t.Fatal("CID indefinido") }
	if c.Version() != 1 { t.Fatalf("versión=%d", c.Version()) }
}
