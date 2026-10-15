package token

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMintAndChain(t *testing.T) {
	dir := t.TempDir()
	r, err := NewRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	aos := `(flow demo (orge x (kind ORGES) (state (x.n 1))))`
	tok, ev, err := r.Mint("process", "caja-rapida", "POS local", "1.0.0", "comercio", aos, nil)
	if err != nil {
		t.Fatal(err)
	}
	if tok.ManifestCID == "" || ev.Hash == "" {
		t.Fatal("missing cid/hash")
	}
	ev2, err := r.Append(tok.ID, "deploy", map[string]any{"app": "caja-rapida"})
	if err != nil {
		t.Fatal(err)
	}
	if ev2.PrevHash != ev.Hash {
		t.Fatalf("chain broken")
	}
	// reload
	r2, err := NewRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, ch, ok := r2.Get(tok.ID)
	if !ok || len(ch) != 2 || got.ChainHead != ev2.Hash {
		t.Fatalf("persist fail %+v %d", got, len(ch))
	}
	if _, err := os.Stat(filepath.Join(dir, tok.ID+".aos")); err != nil {
		t.Fatal(err)
	}
}
