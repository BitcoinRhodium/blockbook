package btr

import (
	"blockbook/bchain"
	"blockbook/bchain/coins/btc"
	"encoding/json"

	"github.com/golang/glog"
)

// BRhodiumRPC is an interface to JSON-RPC BRhodiumNode service.
type BRhodiumRPC struct {
	*btc.BitcoinRPC
}

// NewBCashRPC returns new BRhodiumRPC instance.
func NewBRhodiumRPC(config json.RawMessage, pushHandler func(bchain.NotificationType)) (bchain.BlockChain, error) {
	b, err := btc.NewBitcoinRPC(config, pushHandler)
	if err != nil {
		return nil, err
	}

	s := &BRhodiumRPC{
		b.(*btc.BitcoinRPC),
	}

	s.RPCMarshaler = btc.JSONMarshalerV1{}

	return s, nil
}

// Initialize initializes BRhodiumRPC instance.
func (b *BRhodiumRPC) Initialize() error {
	chainName, err := b.GetChainInfoAndInitializeMempool(b)
	if err != nil {
		return err
	}

	params := GetChainParams(chainName)

	// always create parser
	b.Parser = btc.NewBitcoinParser(params, b.ChainConfig)

	// parameters for getInfo request
	if params.Net == MainnetMagic {
		b.Testnet = false
		b.Network = "livenet"
	} else {
		b.Testnet = true
		b.Network = "testnet"
	}

	glog.Info("rpc: block chain ", params.Name)

	return nil
}
