package xrc
import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/martinboehm/btcd/chaincfg/chainhash"
	"github.com/martinboehm/btcd/wire"
	"github.com/martinboehm/btcutil/chaincfg"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/btc"
	"github.com/trezor/blockbook/bchain/coins/utils"

)

const (
	MainnetMagic wire.BitcoinNet = 0x33333435
	TestnetMagic wire.BitcoinNet = 0x39333435
)

var (
	MainNetParams chaincfg.Params
	TestNetParams chaincfg.Params
)

func init() {
	MainNetParams = chaincfg.MainNetParams
	MainNetParams.Net = MainnetMagic

	// Address encoding magics
	MainNetParams.PubKeyHashAddrID = []byte{61} // base58 prefix: G
	MainNetParams.ScriptHashAddrID = []byte{123} // base58 prefix: A

	TestNetParams = chaincfg.TestNet3Params
	TestNetParams.Net = TestnetMagic

	TestNetParams.PubKeyHashAddrID = []byte{65} // base58 prefix: T
	TestNetParams.ScriptHashAddrID = []byte{128} // base58 prefix: t

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	// see https://github.com/satoshilabs/slips/blob/master/slip-0173.md
	MainNetParams.Bech32HRPSegwit = "xrc"
	TestNetParams.Bech32HRPSegwit = "txrc"
}

// XRhodiumParser handle
type XRhodiumParser struct {
	*btc.BitcoinParser
}

// NewXRhodiumParser returns new XRhodiumParser instance
func NewXRhodiumParser(params *chaincfg.Params, c *btc.Configuration) *XRhodiumParser {
	return &XRhodiumParser{BitcoinParser: btc.NewBitcoinParser(params, c)}
}

// GetChainParams contains network parameters for the main Bitcoin Cash network,
// the regression test Bitcoin Cash network, the test Bitcoin Cash network and
// the simulation test Bitcoin Cash network, in this order
func GetChainParams(chain string) *chaincfg.Params {
	if !chaincfg.IsRegistered(&MainNetParams) {
		err := chaincfg.Register(&MainNetParams)
		if err == nil {
			err = chaincfg.Register(&TestNetParams)
		}
		if err != nil {
			panic(err)
		}
	}
	switch chain {
	case "Test":
		return &TestNetParams
	case "RegTest":
		return &chaincfg.RegressionNetParams
	default:
		return &MainNetParams
	}
}

// headerFixedLength is the length of fixed fields of a block (i.e. without solution)
// see https://github.com/BTCGPU/BTCGPU/wiki/Technical-Spec#block-header
const headerFixedLength = 44 + (chainhash.HashSize * 3)
const timestampOffset = 100
const timestampLength = 4

// ParseBlock parses raw block to our Block struct
func (p *XRhodiumParser) ParseBlock(b []byte) (*bchain.Block, error) {
	r := bytes.NewReader(b)
	time, err := getTimestampAndSkipHeader(r, 0)
	if err != nil {
		return nil, err
	}

	w := wire.MsgBlock{}
	err = utils.DecodeTransactions(r, 0, wire.WitnessEncoding, &w)
	if err != nil {
		return nil, err
	}

	txs := make([]bchain.Tx, len(w.Transactions))
	for ti, t := range w.Transactions {
		txs[ti] = p.TxFromMsgTx(t, false)
	}

	return &bchain.Block{
		BlockHeader: bchain.BlockHeader{
			Size: len(b),
			Time: time,
		},
		Txs: txs,
	}, nil
}

func getTimestampAndSkipHeader(r io.ReadSeeker, pver uint32) (int64, error) {
	_, err := r.Seek(timestampOffset, io.SeekStart)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, timestampLength)
	if _, err = io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	time := binary.LittleEndian.Uint32(buf)

	_, err = r.Seek(headerFixedLength-timestampOffset-timestampLength, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	size, err := wire.ReadVarInt(r, pver)
	if err != nil {
		return 0, err
	}

	_, err = r.Seek(int64(size), io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	return int64(time), nil
}
