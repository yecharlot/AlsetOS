package p2p

import (
	"context"
	"fmt"

	kaddht "github.com/libp2p/go-libp2p-kad-dht"
)

func iniciarDHT(ctx context.Context, nodo *Nodo) error {
	if nodo.DHT != nil {
		return nil
	}
	dht, err := kaddht.New(nodo.Host, nodo.datastore)
	if err != nil {
		return fmt.Errorf("crear DHT: %w", err)
	}
	nodo.DHT = dht
	if err := dht.Bootstrap(ctx); err != nil {
		_ = dht.Close()
		nodo.DHT = nil
		return fmt.Errorf("bootstrap DHT: %w", err)
	}
	return nil
}
