package xrc

import (
	"encoding/json"

	"github.com/golang/glog"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/btc"
)

// BRhodiumRPC is an interface to JSON-RPC BRhodiumNode service.
type XRhodiumRPC struct {
	*btc.BitcoinRPC
}

// NewBCashRPC returns new XRhodiumRPC instance.
func NewXRhodiumRPC(config json.RawMessage, pushHandler func(bchain.NotificationType)) (bchain.BlockChain, error) {
	b, err := btc.NewBitcoinRPC(config, pushHandler)
	if err != nil {
		return nil, err
	}

	s := &XRhodiumRPC{
		b.(*btc.BitcoinRPC),
	}

	s.ParseBlocks = false
	s.RPCMarshaler = btc.JSONMarshalerV1{}
	s.ChainConfig.SupportsEstimateSmartFee = false

	return s, nil
}

// Initialize initializes XRhodiumRPC instance.
func (b *XRhodiumRPC) Initialize() error {
	ci, err := b.GetChainInfo()
	if err != nil {
		return err
	}
	chainName := ci.Chain

	params := GetChainParams(chainName)

	// always create parser
	b.Parser = NewXRhodiumParser(params, b.ChainConfig)

	// parameters for getInfo request
	if params.Net == MainnetMagic {
		b.Testnet = false
		b.Network = "mainnet"
	} else {
		b.Testnet = true
		b.Network = "testnet"
	}

	glog.Info("rpc: block chain ", params.Name)

	return nil
}
