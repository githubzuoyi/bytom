package instance

import (
	"github.com/bytom/encoding/json"
	"github.com/bytom/protocol/bc"
	"github.com/bytom/crypto/ed25519"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/bytom/equity/compiler"
	"github.com/bytom/protocol/vm"
)

// CallOptionBodyBytes refer to contract's body
var CallOptionBodyBytes []byte

func init() {
	CallOptionBodyBytes, _ = hex.DecodeString("557a6420000000547acda069547a547aae7cac69007c7b51547ac1632c000000547acd9f6900c3c251567ac1")
}

// contract CallOption(strikePrice: Amount, strikeCurrency: Asset, seller: Program, buyerKey: PublicKey, blockHeight: Integer) locks underlying
//
// 5                        [... <clause selector> blockHeight buyerKey seller strikeCurrency strikePrice 5]
// ROLL                     [... blockHeight buyerKey seller strikeCurrency strikePrice <clause selector>]
// JUMPIF:$expire           [... blockHeight buyerKey seller strikeCurrency strikePrice]
// $exercise                [... blockHeight buyerKey seller strikeCurrency strikePrice]
// 4                        [... buyerSig blockHeight buyerKey seller strikeCurrency strikePrice 4]
// ROLL                     [... buyerSig buyerKey seller strikeCurrency strikePrice blockHeight]
// BLOCKHEIGHT GREATERTHAN  [... buyerSig buyerKey seller strikeCurrency strikePrice below(blockHeight)]
// VERIFY                   [... buyerSig buyerKey seller strikeCurrency strikePrice]
// 4                        [... buyerSig buyerKey seller strikeCurrency strikePrice 4]
// ROLL                     [... buyerKey seller strikeCurrency strikePrice buyerSig]
// 4                        [... buyerKey seller strikeCurrency strikePrice buyerSig 4]
// ROLL                     [... seller strikeCurrency strikePrice buyerSig buyerKey]
// TXSIGHASH SWAP CHECKSIG  [... seller strikeCurrency strikePrice checkTxSig(buyerKey, buyerSig)]
// VERIFY                   [... seller strikeCurrency strikePrice]
// 0                        [... seller strikeCurrency strikePrice 0]
// SWAP                     [... seller strikeCurrency 0 strikePrice]
// 2                        [... seller strikeCurrency 0 strikePrice 2]
// ROLL                     [... seller 0 strikePrice strikeCurrency]
// 1                        [... seller 0 strikePrice strikeCurrency 1]
// 4                        [... seller 0 strikePrice strikeCurrency 1 4]
// ROLL                     [... 0 strikePrice strikeCurrency 1 seller]
// CHECKOUTPUT              [... checkOutput(payment, seller)]
// JUMP:$_end               [... blockHeight buyerKey seller strikeCurrency strikePrice]
// $expire                  [... blockHeight buyerKey seller strikeCurrency strikePrice]
// 4                        [... blockHeight buyerKey seller strikeCurrency strikePrice 4]
// ROLL                     [... buyerKey seller strikeCurrency strikePrice blockHeight]
// BLOCKHEIGHT LESSTHAN     [... buyerKey seller strikeCurrency strikePrice above(blockHeight)]
// VERIFY                   [... buyerKey seller strikeCurrency strikePrice]
// 0                        [... buyerKey seller strikeCurrency strikePrice 0]
// AMOUNT                   [... buyerKey seller strikeCurrency strikePrice 0 <amount>]
// ASSET                    [... buyerKey seller strikeCurrency strikePrice 0 <amount> <asset>]
// 1                        [... buyerKey seller strikeCurrency strikePrice 0 <amount> <asset> 1]
// 6                        [... buyerKey seller strikeCurrency strikePrice 0 <amount> <asset> 1 6]
// ROLL                     [... buyerKey strikeCurrency strikePrice 0 <amount> <asset> 1 seller]
// CHECKOUTPUT              [... buyerKey strikeCurrency strikePrice checkOutput(underlying, seller)]
// $_end                    [... blockHeight buyerKey seller strikeCurrency strikePrice]

// PayToCallOption instantiates contract CallOption as a program with specific arguments.
func PayToCallOption(strikePrice uint64, strikeCurrency bc.AssetID, seller []byte, buyerKey ed25519.PublicKey, blockHeight int64) ([]byte, error) {
	_contractParams := []*compiler.Param{
		{Name: "strikePrice", Type: "Amount"},
		{Name: "strikeCurrency", Type: "Asset"},
		{Name: "seller", Type: "Program"},
		{Name: "buyerKey", Type: "PublicKey"},
		{Name: "blockHeight", Type: "Integer"},
	}
	var _contractArgs []compiler.ContractArg
	_strikePrice := int64(strikePrice)
	_contractArgs = append(_contractArgs, compiler.ContractArg{I: &_strikePrice})
	_strikeCurrency := strikeCurrency.Bytes()
	_contractArgs = append(_contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&_strikeCurrency)})
	_contractArgs = append(_contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&seller)})
	_contractArgs = append(_contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&buyerKey)})
	_contractArgs = append(_contractArgs, compiler.ContractArg{I: &blockHeight})
	return compiler.Instantiate(CallOptionBodyBytes, _contractParams, false, _contractArgs)
}

// ParsePayToCallOption parses the arguments out of an instantiation of contract CallOption.
// If the input is not an instantiation of CallOption, returns an error.
func ParsePayToCallOption(prog []byte) ([][]byte, error) {
	var result [][]byte
	insts, err := vm.ParseProgram(prog)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 5; i++ {
		if len(insts) == 0 {
			return nil, fmt.Errorf("program too short")
		}
		if !insts[0].IsPushdata() {
			return nil, fmt.Errorf("too few arguments")
		}
		result = append(result, insts[0].Data)
		insts = insts[1:]
	}
	if len(insts) != 4 {
		return nil, fmt.Errorf("program too short")
	}
	if insts[0].Op != vm.OP_DEPTH {
		return nil, fmt.Errorf("wrong program format")
	}
	if !insts[1].IsPushdata() {
		return nil, fmt.Errorf("wrong program format")
	}
	if !bytes.Equal(CallOptionBodyBytes, insts[1].Data) {
		return nil, fmt.Errorf("body bytes do not match CallOption")
	}
	if !insts[2].IsPushdata() {
		return nil, fmt.Errorf("wrong program format")
	}
	v, err := vm.AsInt64(insts[2].Data)
	if err != nil {
		return nil, err
	}
	if v != 0 {
		return nil, fmt.Errorf("wrong program format")
	}
	if insts[3].Op != vm.OP_CHECKPREDICATE {
		return nil, fmt.Errorf("wrong program format")
	}
	return result, nil
}

