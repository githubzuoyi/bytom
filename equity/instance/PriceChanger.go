package instance

import (
	"github.com/bytom/protocol/vm"
	"github.com/bytom/encoding/json"
	"github.com/bytom/protocol/bc"
	"github.com/bytom/crypto/ed25519"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/bytom/equity/compiler"
)

// PriceChangerBodyBytes refer to contract's body
var PriceChangerBodyBytes []byte

func init() {
	PriceChangerBodyBytes, _ = hex.DecodeString("557a6432000000557a5479ae7cac6900c3c25100597a89587a89587a89587a89557a890274787e008901c07ec1633a000000007b537a51567ac1")
}

// contract PriceChanger(askAmount: Amount, askAsset: Asset, sellerKey: PublicKey, sellerProg: Program) locks offered
//
// 5                        [... <clause selector> sellerProg sellerKey askAsset askAmount PriceChanger 5]
// ROLL                     [... sellerProg sellerKey askAsset askAmount PriceChanger <clause selector>]
// JUMPIF:$redeem           [... sellerProg sellerKey askAsset askAmount PriceChanger]
// $changePrice             [... sellerProg sellerKey askAsset askAmount PriceChanger]
// 5                        [... newAmount newAsset sig sellerProg sellerKey askAsset askAmount PriceChanger 5]
// ROLL                     [... newAmount newAsset sellerProg sellerKey askAsset askAmount PriceChanger sig]
// 4                        [... newAmount newAsset sellerProg sellerKey askAsset askAmount PriceChanger sig 4]
// PICK                     [... newAmount newAsset sellerProg sellerKey askAsset askAmount PriceChanger sig sellerKey]
// TXSIGHASH SWAP CHECKSIG  [... newAmount newAsset sellerProg sellerKey askAsset askAmount PriceChanger checkTxSig(sellerKey, sig)]
// VERIFY                   [... newAmount newAsset sellerProg sellerKey askAsset askAmount PriceChanger]
// 0                        [... newAmount newAsset sellerProg sellerKey askAsset askAmount PriceChanger 0]
// AMOUNT                   [... newAmount newAsset sellerProg sellerKey askAsset askAmount PriceChanger 0 <amount>]
// ASSET                    [... newAmount newAsset sellerProg sellerKey askAsset askAmount PriceChanger 0 <amount> <asset>]
// 1                        [... newAmount newAsset sellerProg sellerKey askAsset askAmount PriceChanger 0 <amount> <asset> 1]
// 0                        [... newAmount newAsset sellerProg sellerKey askAsset askAmount PriceChanger 0 <amount> <asset> 1 0]
// 9                        [... newAmount newAsset sellerProg sellerKey askAsset askAmount PriceChanger 0 <amount> <asset> 1 0 9]
// ROLL                     [... newAmount newAsset sellerKey askAsset askAmount PriceChanger 0 <amount> <asset> 1 0 sellerProg]
// CATPUSHDATA              [... newAmount newAsset sellerKey askAsset askAmount PriceChanger 0 <amount> <asset> 1 PriceChanger(...)]
// 8                        [... newAmount newAsset sellerKey askAsset askAmount PriceChanger 0 <amount> <asset> 1 PriceChanger(...) 8]
// ROLL                     [... newAmount newAsset askAsset askAmount PriceChanger 0 <amount> <asset> 1 PriceChanger(...) sellerKey]
// CATPUSHDATA              [... newAmount newAsset askAsset askAmount PriceChanger 0 <amount> <asset> 1 PriceChanger(...)]
// 8                        [... newAmount newAsset askAsset askAmount PriceChanger 0 <amount> <asset> 1 PriceChanger(...) 8]
// ROLL                     [... newAmount askAsset askAmount PriceChanger 0 <amount> <asset> 1 PriceChanger(...) newAsset]
// CATPUSHDATA              [... newAmount askAsset askAmount PriceChanger 0 <amount> <asset> 1 PriceChanger(...)]
// 8                        [... newAmount askAsset askAmount PriceChanger 0 <amount> <asset> 1 PriceChanger(...) 8]
// ROLL                     [... askAsset askAmount PriceChanger 0 <amount> <asset> 1 PriceChanger(...) newAmount]
// CATPUSHDATA              [... askAsset askAmount PriceChanger 0 <amount> <asset> 1 PriceChanger(...)]
// 5                        [... askAsset askAmount PriceChanger 0 <amount> <asset> 1 PriceChanger(...) 5]
// ROLL                     [... askAsset askAmount 0 <amount> <asset> 1 PriceChanger(...) PriceChanger]
// CATPUSHDATA              [... askAsset askAmount 0 <amount> <asset> 1 PriceChanger(...)]
// 0x7478                   [... askAsset askAmount 0 <amount> <asset> 1 PriceChanger(...) 0x7478]
// CAT                      [... askAsset askAmount 0 <amount> <asset> 1 PriceChanger(...)]
// 0                        [... askAsset askAmount 0 <amount> <asset> 1 PriceChanger(...) 0]
// CATPUSHDATA              [... askAsset askAmount 0 <amount> <asset> 1 PriceChanger(...)]
// 192                      [... askAsset askAmount 0 <amount> <asset> 1 PriceChanger(...) 192]
// CAT                      [... askAsset askAmount 0 <amount> <asset> 1 PriceChanger(newAmount, newAsset, sellerKey, sellerProg)]
// CHECKOUTPUT              [... askAsset askAmount checkOutput(offered, PriceChanger(newAmount, newAsset, sellerKey, sellerProg))]
// JUMP:$_end               [... sellerProg sellerKey askAsset askAmount PriceChanger]
// $redeem                  [... sellerProg sellerKey askAsset askAmount PriceChanger]
// 0                        [... sellerProg sellerKey askAsset askAmount PriceChanger 0]
// 2                        [... sellerProg sellerKey askAsset askAmount PriceChanger 0 2]
// ROLL                     [... sellerProg sellerKey askAsset PriceChanger 0 askAmount]
// 3                        [... sellerProg sellerKey askAsset PriceChanger 0 askAmount 3]
// ROLL                     [... sellerProg sellerKey PriceChanger 0 askAmount askAsset]
// 1                        [... sellerProg sellerKey PriceChanger 0 askAmount askAsset 1]
// 6                        [... sellerProg sellerKey PriceChanger 0 askAmount askAsset 1 6]
// ROLL                     [... sellerKey PriceChanger 0 askAmount askAsset 1 sellerProg]
// CHECKOUTPUT              [... sellerKey PriceChanger checkOutput(payment, sellerProg)]
// $_end                    [... sellerProg sellerKey askAsset askAmount PriceChanger]

// PayToPriceChanger instantiates contract PriceChanger as a program with specific arguments.
func PayToPriceChanger(askAmount uint64, askAsset bc.AssetID, sellerKey ed25519.PublicKey, sellerProg []byte) ([]byte, error) {
	_contractParams := []*compiler.Param{
		{Name: "askAmount", Type: "Amount"},
		{Name: "askAsset", Type: "Asset"},
		{Name: "sellerKey", Type: "PublicKey"},
		{Name: "sellerProg", Type: "Program"},
	}
	var _contractArgs []compiler.ContractArg
	_askAmount := int64(askAmount)
	_contractArgs = append(_contractArgs, compiler.ContractArg{I: &_askAmount})
	_askAsset := askAsset.Bytes()
	_contractArgs = append(_contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&_askAsset)})
	_contractArgs = append(_contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&sellerKey)})
	_contractArgs = append(_contractArgs, compiler.ContractArg{S: (*json.HexBytes)(&sellerProg)})
	return compiler.Instantiate(PriceChangerBodyBytes, _contractParams, true, _contractArgs)
}

// ParsePayToPriceChanger parses the arguments out of an instantiation of contract PriceChanger.
// If the input is not an instantiation of PriceChanger, returns an error.
func ParsePayToPriceChanger(prog []byte) ([][]byte, error) {
	var result [][]byte
	insts, err := vm.ParseProgram(prog)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 4; i++ {
		if len(insts) == 0 {
			return nil, fmt.Errorf("program too short")
		}
		if !insts[0].IsPushdata() {
			return nil, fmt.Errorf("too few arguments")
		}
		result = append(result, insts[0].Data)
		insts = insts[1:]
	}
	if len(insts) == 0 {
		return nil, fmt.Errorf("program too short")
	}
	if !insts[0].IsPushdata() {
		return nil, fmt.Errorf("too few arguments")
	}
	if !bytes.Equal(PriceChangerBodyBytes, insts[0].Data) {
		return nil, fmt.Errorf("body bytes do not match PriceChanger")
	}
	insts = insts[1:]
	if len(insts) != 4 {
		return nil, fmt.Errorf("program too short")
	}
	if insts[0].Op != vm.OP_DEPTH {
		return nil, fmt.Errorf("wrong program format")
	}
	if insts[1].Op != vm.OP_OVER {
		return nil, fmt.Errorf("wrong program format")
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

