/*
* Copyright (C) 2020 The poly network Authors
* This file is part of The poly network library.
*
* The poly network is free software: you can redistribute it and/or modify
* it under the terms of the GNU Lesser General Public License as published by
* the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* The poly network is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
* GNU Lesser General Public License for more details.
* You should have received a copy of the GNU Lesser General Public License
* along with The poly network . If not, see <http://www.gnu.org/licenses/>.
 */
package lib

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"strings"
	"sync"

	types2 "github.com/cosmos/cosmos-sdk/types"
	"github.com/joeqian10/neo-gogogo/block"
	"github.com/joeqian10/neo-gogogo/helper/io"
	"github.com/joeqian10/neo-gogogo/rpc"
	"github.com/ontio/ontology-crypto/keypair"
	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
	"github.com/polynetwork/cosmos-poly-module/headersync"
	poly_go_sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/password"
	vconfig "github.com/polynetwork/poly/consensus/vbft/config"
	"github.com/polynetwork/poly/core/payload"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/header_sync/bsc"
	"github.com/polynetwork/poly/native/service/header_sync/cosmos"
	"github.com/polynetwork/poly/native/states"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/rpc/client/http"
)

var wg sync.WaitGroup

func RegisterCandidate(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]
	pk, err := vconfig.Pubkey(args[0])
	if err != nil {
		return err
	}
	candidate := types.AddressFromPubKey(pk)
	txHash, err := poly.Native.Nm.RegisterCandidate(args[0], acc)
	if err != nil {
		if strings.Contains(err.Error(), "already") {
			fmt.Printf("candidate %s already registered: %v", acc.Address.ToBase58(), err)
			return nil
		}
		return fmt.Errorf("sendTransaction error: %v", err)
	}

	WaitPolyTx(txHash, poly)
	fmt.Printf("successful to register candidate: ( candidate: %s, txhash: %s, signer: %s )\n",
		candidate.ToBase58(), txHash.ToHexString(), acc.Address.ToBase58())

	return nil
}

func ApproveCandidate(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	pk, err := vconfig.Pubkey(args[0])
	if err != nil {
		return err
	}
	candidate := types.AddressFromPubKey(pk)
	wg.Add(len(accs))
	for i, acc := range accs {
		go func(acc *poly_go_sdk.Account) {
			defer wg.Done()
			txhash, err := poly.Native.Nm.ApproveCandidate(args[i], acc)
			if err != nil {
				fmt.Printf("sendTransaction error: %v", err)
				return
			}
			WaitPolyTx(txhash, poly)
			fmt.Printf("successful to approve candidate: ( candidate: %s, txhash: %s, acc: %s )\n",
				candidate.ToBase58(), txhash.ToHexString(), acc.Address.ToBase58())
		}(acc)
	}
	wg.Wait()

	return nil
}

func UnRegisterCandidate(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]

	txHash, err := poly.Native.Nm.UnRegisterCandidate(vconfig.PubkeyID(acc.PublicKey), acc)
	if err != nil {
		return fmt.Errorf("sendTransaction error: %v", err)
	}

	WaitPolyTx(txHash, poly)
	fmt.Printf("successful to unregister candidate: ( candidate: %s, txhash: %s, signer: %s )\n",
		acc.Address.ToBase58(), txHash.ToHexString(), acc.Address.ToBase58())

	return nil
}

func QuitNode(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]

	txhash, err := poly.Native.Nm.QuitNode(vconfig.PubkeyID(acc.PublicKey), acc)
	if err != nil {
		return fmt.Errorf("failed to quit %s: %v", acc.Address.ToBase58(), err)
	}
	WaitPolyTx(txhash, poly)
	fmt.Printf("successful to quit node %s on Poly: txhash: %s\n", acc.Address.ToBase58(), txhash.ToHexString())

	return nil
}

func BlackNode(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}

	pk, err := vconfig.Pubkey(args[0])
	if err != nil {
		return err
	}
	node := types.AddressFromPubKey(pk)
	wg.Add(len(accs))
	for _, acc := range accs {
		go func(acc *poly_go_sdk.Account) {
			defer wg.Done()
			txhash, err := poly.Native.Nm.BlackNode([]string{args[0]}, acc)
			if err != nil {
				fmt.Printf("failed to black %s: %v", acc.Address.ToBase58(), err)
				return
			}
			WaitPolyTx(txhash, poly)
			fmt.Printf("successful to vote yes to black node %s on Poly: txhash: %s\n", node.ToBase58(), txhash.ToHexString())
		}(acc)
	}
	wg.Wait()

	return nil
}

func WhiteNode(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	pk, err := vconfig.Pubkey(args[0])
	if err != nil {
		return err
	}
	node := types.AddressFromPubKey(pk)
	wg.Add(len(accs))
	for _, acc := range accs {
		go func(acc *poly_go_sdk.Account) {
			defer wg.Done()
			txhash, err := poly.Native.Nm.BlackNode([]string{args[0]}, acc)
			if err != nil {
				fmt.Printf("failed to white %s: %v\n", acc.Address.ToBase58(), err)
				return
			}
			WaitPolyTx(txhash, poly)
			fmt.Printf("successful to vote yes to white node %s on Poly: txhash: %s\n", node.ToBase58(), txhash.ToHexString())
		}(acc)
	}
	wg.Wait()

	return nil
}

func CreateCommitDposTx(cmd *cobra.Command, args []string) error {
	poly := poly_go_sdk.NewPolySdk()
	tx, err := poly.Native.Nm.NewCommitDposTransaction()
	if err != nil {
		return err
	}

	str, err := cmd.Flags().GetString(ConsensusPubKeys)
	pks := strings.Split(str, ",")
	if err != nil {
		return err
	}
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for i, v := range pks {
		pk, err := vconfig.Pubkey(v)
		if err != nil {
			return fmt.Errorf("failed to get no%d pubkey: %v", i, err)
		}
		pubKeys = append(pubKeys, pk)
	}

	tx.Sigs = append(tx.Sigs, types.Sig{
		SigData: make([][]byte, 0),
		M:       uint16(len(pubKeys) - (len(pubKeys)-1)/3),
		PubKeys: pubKeys,
	})
	sink := common.NewZeroCopySink(nil)
	if err := tx.Serialization(sink); err != nil {
		return err
	}

	fmt.Printf("raw transaction is %s\nNeed to send this transaction to every single consensus peer to sign. \n",
		hex.EncodeToString(sink.Bytes()))
	return nil
}

func SignCommitDposTx(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]

	tx := &types.Transaction{}
	raw, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}
	if err := tx.Deserialization(common.NewZeroCopySource(raw)); err != nil {
		return err
	}
	if err = poly.MultiSignToTransaction(tx, tx.Sigs[0].M, tx.Sigs[0].PubKeys, acc); err != nil {
		return fmt.Errorf("multi sign failed, err: %s", err)
	}

	sink := common.NewZeroCopySink(nil)
	err = tx.Serialization(sink)
	if err != nil {
		return err
	}

	if uint16(len(tx.Sigs[0].SigData)) >= tx.Sigs[0].M {
		txhash, err := poly.SendTransaction(tx)
		if err != nil {
			return fmt.Errorf("failed to send tx to poly: %v\nRaw Tx: %s\n", err, hex.EncodeToString(sink.Bytes()))
		}
		WaitPolyTx(txhash, poly)
		fmt.Printf("successful to sign commit DPOS tx and send tx %s to Poly with enough sigs\n", txhash.ToHexString())
		return nil
	}
	fmt.Printf("successful to sign tx and %d/%d sigs now, need at least %d sig: raw tx: %s\n", len(tx.Sigs[0].SigData),
		len(tx.Sigs[0].PubKeys), tx.Sigs[0].M, hex.EncodeToString(sink.Bytes()))

	return nil
}

func CreateUpdateConfigTx(cmd *cobra.Command, args []string) error {
	blockMsgDelay, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return err
	}
	hashMsgDelay, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		return err
	}
	peerHandshakeTimeout, err := strconv.ParseUint(args[2], 10, 32)
	if err != nil {
		return err
	}
	maxBlockChangeView, err := strconv.ParseUint(args[3], 10, 32)
	if err != nil {
		return err
	}

	poly := poly_go_sdk.NewPolySdk()
	tx, err := poly.Native.Nm.NewUpdateConfigTransaction(uint32(blockMsgDelay), uint32(hashMsgDelay),
		uint32(peerHandshakeTimeout), uint32(maxBlockChangeView))
	if err != nil {
		return err
	}

	str, err := cmd.Flags().GetString(ConsensusPubKeys)
	pks := strings.Split(str, ",")
	if err != nil {
		return err
	}
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for i, v := range pks {
		pk, err := vconfig.Pubkey(v)
		if err != nil {
			return fmt.Errorf("failed to get no%d pubkey: %v", i, err)
		}
		pubKeys = append(pubKeys, pk)
	}

	tx.Sigs = append(tx.Sigs, types.Sig{
		SigData: make([][]byte, 0),
		M:       uint16(len(pubKeys) - (len(pubKeys)-1)/3),
		PubKeys: pubKeys,
	})
	sink := common.NewZeroCopySink(nil)
	if err := tx.Serialization(sink); err != nil {
		return err
	}

	fmt.Printf("raw transaction is %s\nNeed to send this transaction to every single consensus peer to sign. \n",
		hex.EncodeToString(sink.Bytes()))
	return nil
}

func SignUpdateConfigTx(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]

	tx := &types.Transaction{}
	raw, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}
	if err := tx.Deserialization(common.NewZeroCopySource(raw)); err != nil {
		return err
	}

	code := tx.Payload.(*payload.InvokeCode)
	param := &states.ContractInvokeParam{}
	if err := param.Deserialization(common.NewZeroCopySource(code.Code)); err != nil {
		return err
	}
	np := &node_manager.UpdateConfigParam{}
	if err := np.Deserialization(common.NewZeroCopySource(param.Args)); err != nil {
		return err
	}
	fmt.Printf("New configuration in this transaction: ( blockMsgDelay: %d, hashMsgDelay: %d, "+
		"peerHandshakeTimeout: %d, maxBlockChangeView: %d )\n", np.Configuration.BlockMsgDelay,
		np.Configuration.HashMsgDelay, np.Configuration.PeerHandshakeTimeout, np.Configuration.MaxBlockChangeView)

	if err = poly.MultiSignToTransaction(tx, tx.Sigs[0].M, tx.Sigs[0].PubKeys, acc); err != nil {
		return fmt.Errorf("multi sign failed, err: %s", err)
	}

	sink := common.NewZeroCopySink(nil)
	err = tx.Serialization(sink)
	if err != nil {
		return err
	}

	if uint16(len(tx.Sigs[0].SigData)) >= tx.Sigs[0].M {
		txhash, err := poly.SendTransaction(tx)
		if err != nil {
			return fmt.Errorf("failed to send tx to poly: %v\nRaw Tx: %s\n", err, hex.EncodeToString(sink.Bytes()))
		}
		WaitPolyTx(txhash, poly)
		fmt.Printf("successful to sign update config tx and send tx %s to Poly with enough sigs\n", txhash.ToHexString())
		return nil
	}
	fmt.Printf("successful to sign tx and %d/%d sigs now, need at least %d sig: raw tx: %s\n", len(tx.Sigs[0].SigData),
		len(tx.Sigs[0].PubKeys), tx.Sigs[0].M, hex.EncodeToString(sink.Bytes()))

	return nil
}

func RegisterRelayer(cmd *cobra.Command, args []string) error {
	var err error
	addrs := make([]common.Address, len(args))
	for i, v := range args {
		addrs[i], err = common.AddressFromBase58(v)
		if err != nil {
			return fmt.Errorf("no%d address decode failed: %v", i, err)
		}
	}

	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]
	txhash, err := poly.Native.Rm.RegisterRelayer(addrs, acc)
	if err != nil {
		return err
	}
	WaitPolyTx(txhash, poly)
	event, err := poly.GetSmartContractEvent(txhash.ToHexString())
	if err != nil {
		return err
	}
	var id uint64
	for _, e := range event.Notify {
		states := e.States.([]interface{})
		if states[0].(string) == "putRelayerApply" {
			id = uint64(states[1].(float64))
		}
	}
	fmt.Printf("successful to register %v, and id is %d: txhash: %s\n", args, id, txhash.ToHexString())

	return nil
}

func ApproveRegisterRelayer(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	wg.Add(len(accs))
	for _, acc := range accs {
		go func(acc *poly_go_sdk.Account) {
			defer wg.Done()
			txhash, err := poly.Native.Rm.ApproveRegisterRelayer(id, acc)
			if err != nil {
				fmt.Printf("err approving with acc: %s: %v\n", acc.Address.ToBase58(), err)
				return
			}
			WaitPolyTx(txhash, poly)
			fmt.Printf("successful to approve registration id %d: txhash: %s\n", id, txhash.ToHexString())
		}(acc)
	}
	wg.Wait()

	return nil
}

func RemoveRelayer(cmd *cobra.Command, args []string) error {
	var err error
	addrs := make([]common.Address, len(args))
	for i, v := range args {
		addrs[i], err = common.AddressFromBase58(v)
		if err != nil {
			return fmt.Errorf("no%d address decode failed: %v", i, err)
		}
	}

	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]
	txhash, err := poly.Native.Rm.RemoveRelayer(addrs, acc)
	if err != nil {
		return err
	}
	WaitPolyTx(txhash, poly)
	event, err := poly.GetSmartContractEvent(txhash.ToHexString())
	if err != nil {
		return err
	}
	var id uint64
	for _, e := range event.Notify {
		states := e.States.([]interface{})
		if states[0].(string) == "putRelayerRemove" {
			id = uint64(states[1].(float64))
		}
	}
	fmt.Printf("successful to remove relayers %v, and id is %d: txhash: %s\n", args, id, txhash.ToHexString())

	return nil
}

func ApproveRemoveRelayer(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	wg.Add(len(accs))
	for _, acc := range accs {
		go func(acc *poly_go_sdk.Account) {
			defer wg.Done()
			txhash, err := poly.Native.Rm.ApproveRemoveRelayer(id, acc)
			if err != nil {
				fmt.Printf("err approving with acc: %s: %v\n", acc.Address.ToBase58(), err)
				return
			}
			WaitPolyTx(txhash, poly)
			fmt.Printf("successful to approve remove id %d: txhash: %s\n", id, txhash.ToHexString())
		}(acc)
	}
	wg.Wait()

	return nil
}

func RegisterSideChain(cmd *cobra.Command, args []string) error {
	chainId, err := cmd.Flags().GetUint64(ChainId)
	if err != nil {
		return err
	}
	router, err := cmd.Flags().GetUint64(Router)
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString(Name)
	if err != nil {
		return err
	}
	num, err := cmd.Flags().GetUint64(BlkToWait)
	if err != nil {
		return err
	}
	cmcc, err := cmd.Flags().GetString(CMCC)
	if err != nil {
		return err
	}
	extra, err := cmd.Flags().GetString(ExtraInfo)
	if err != nil {
		return err
	}

	cmcc = strings.TrimPrefix(cmcc, "0x")
	var cmccAddr []byte
	if cmcc == "" {
		cmccAddr = []byte{}
	} else {
		cmccAddr, err = hex.DecodeString(cmcc)
		if err != nil {
			return err
		}
	}
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]

	var txhash common.Uint256
	if extra == "" {
		txhash, err = poly.Native.Scm.RegisterSideChain(acc.Address, chainId, router, name, num, cmccAddr, acc)
	} else {
		txhash, err = poly.Native.Scm.RegisterSideChainExt(acc.Address, chainId, router, name, num, cmccAddr, []byte(extra), acc)
	}
	if err != nil {
		return err
	}
	WaitPolyTx(txhash, poly)
	fmt.Printf("successful to register side chain: txhash: %s\n", txhash.ToHexString())

	return nil
}

func ApproveRegisterSideChain(cmd *cobra.Command, args []string) error {
	chainId, err := cmd.Flags().GetUint64(ChainId)
	if err != nil {
		return err
	}
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	wg.Add(len(accs))
	for _, acc := range accs {
		go func(acc *poly_go_sdk.Account) {
			defer wg.Done()
			txhash, err := poly.Native.Scm.ApproveRegisterSideChain(chainId, acc)
			if err != nil {
				fmt.Printf("ApproveRegisterSideChain failed with acc %s: %v\n", acc.Address.ToBase58(), err)
				return
			}
			WaitPolyTx(txhash, poly)
			fmt.Printf("successful to approve: ( acc: %s, txhash: %s, chain-id: %d )\n",
				acc.Address.ToBase58(), txhash.ToHexString(), chainId)
		}(acc)
	}
	wg.Wait()

	return nil
}

func UpdateSideChain(cmd *cobra.Command, args []string) error {
	chainId, err := cmd.Flags().GetUint64(ChainId)
	if err != nil {
		return err
	}
	router, err := cmd.Flags().GetUint64(Router)
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString(Name)
	if err != nil {
		return err
	}
	num, err := cmd.Flags().GetUint64(BlkToWait)
	if err != nil {
		return err
	}
	cmcc, err := cmd.Flags().GetString(CMCC)
	if err != nil {
		return err
	}
	extra, err := cmd.Flags().GetString(ExtraInfo)
	if err != nil {
		return err
	}

	cmcc = strings.TrimPrefix(cmcc, "0x")
	var cmccAddr []byte
	if cmcc == "" {
		cmccAddr = []byte{}
	} else {
		cmccAddr, err = hex.DecodeString(cmcc)
		if err != nil {
			return err
		}
	}
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]

	var txhash common.Uint256
	if extra == "" {
		txhash, err = poly.Native.Scm.UpdateSideChain(acc.Address, chainId, router, name, num, cmccAddr, acc)
	} else {
		txhash, err = poly.Native.Scm.UpdateSideChainExt(acc.Address, chainId, router, name, num, cmccAddr, []byte(extra), acc)
	}
	if err != nil {
		return err
	}
	WaitPolyTx(txhash, poly)
	fmt.Printf("successful to update side chain: txhash: %s\n", txhash.ToHexString())

	return nil
}

func ApproveUpdateSideChain(cmd *cobra.Command, args []string) error {
	chainId, err := cmd.Flags().GetUint64(ChainId)
	if err != nil {
		return err
	}
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	wg.Add(len(accs))
	for _, acc := range accs {
		go func(acc *poly_go_sdk.Account) {
			defer wg.Done()
			txhash, err := poly.Native.Scm.ApproveUpdateSideChain(chainId, acc)
			if err != nil {
				fmt.Printf("ApproveUpdateSideChain failed with acc: %s: %v\n", acc.Address.ToBase58(), err)
				return
			}
			fmt.Printf("successful to approve: ( acc: %s, txhash: %s, chain-id: %d )\n",
				acc.Address.ToBase58(), txhash.ToHexString(), chainId)
		}(acc)
	}
	wg.Wait()

	return nil
}

func QuitSideChain(cmd *cobra.Command, args []string) error {
	chainId, err := cmd.Flags().GetUint64(ChainId)
	if err != nil {
		return err
	}
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]
	txhash, err := poly.Native.Scm.QuitSideChain(chainId, acc)
	if err != nil {
		return fmt.Errorf("QuitSideChain failed: %v", err)
	}
	fmt.Printf("successful to quit chain: ( acc: %s, txhash: %s, chain-id: %d )\n",
		acc.Address.ToBase58(), txhash.ToHexString(), chainId)

	return nil
}

func ApproveQuitSideChain(cmd *cobra.Command, args []string) error {
	chainId, err := cmd.Flags().GetUint64(ChainId)
	if err != nil {
		return err
	}
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	wg.Add(len(accs))
	for _, acc := range accs {
		go func(acc *poly_go_sdk.Account) {
			defer wg.Done()
			txhash, err := poly.Native.Scm.ApproveQuitSideChain(chainId, acc)
			if err != nil {
				fmt.Printf("ApproveQuitSideChain failed with acc: %s: %v\n", acc.Address.ToBase58(), err)
				return
			}
			fmt.Printf("successful to approve quit chain: ( acc: %s, txhash: %s, chain-id: %d )\n",
				acc.Address.ToBase58(), txhash.ToHexString(), chainId)
		}(acc)
	}
	wg.Wait()

	return nil
}

func CreateSyncOntGenesisHdrToPolyTx(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	h, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	ontRpc, err := cmd.Flags().GetString(OntRpcAddr)
	if err != nil {
		return err
	}
	ont := ontology_go_sdk.NewOntologySdk()
	ont.NewRpcClient().SetAddress(ontRpc)
	blk, err := ont.GetBlockByHeight(uint32(h))
	if err != nil {
		return err
	}

	poly := poly_go_sdk.NewPolySdk()
	tx, err := poly.Native.Hs.NewSyncGenesisHeaderTransaction(id, blk.Header.ToArray())
	if err != nil {
		return err
	}

	str, err := cmd.Flags().GetString(ConsensusPubKeys)
	pks := strings.Split(str, ",")
	if err != nil {
		return err
	}
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for i, v := range pks {
		pk, err := vconfig.Pubkey(v)
		if err != nil {
			return fmt.Errorf("failed to get no%d pubkey: %v", i, err)
		}
		pubKeys = append(pubKeys, pk)
	}

	tx.Sigs = append(tx.Sigs, types.Sig{
		SigData: make([][]byte, 0),
		M:       uint16(len(pubKeys) - (len(pubKeys)-1)/3),
		PubKeys: pubKeys,
	})
	sink := common.NewZeroCopySink(nil)
	if err := tx.Serialization(sink); err != nil {
		return err
	}

	fmt.Printf("raw transaction is %s\nNeed to send this transaction to every single consensus peer to sign. \n",
		hex.EncodeToString(sink.Bytes()))
	return nil
}

func CreateSyncEthGenesisHdrToPolyTx(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	h, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	ethRpc, err := cmd.Flags().GetString(EthRpcAddr)
	if err != nil {
		return err
	}
	et := NewEthTools(ethRpc)
	hdr, err := et.GetBlockHeader(h)
	if err != nil {
		return err
	}
	raw, err := hdr.MarshalJSON()
	if err != nil {
		return err
	}

	poly := poly_go_sdk.NewPolySdk()
	tx, err := poly.Native.Hs.NewSyncGenesisHeaderTransaction(id, raw)
	if err != nil {
		return err
	}

	str, err := cmd.Flags().GetString(ConsensusPubKeys)
	pks := strings.Split(str, ",")
	if err != nil {
		return err
	}
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for i, v := range pks {
		pk, err := vconfig.Pubkey(v)
		if err != nil {
			return fmt.Errorf("failed to get no%d pubkey: %v", i, err)
		}
		pubKeys = append(pubKeys, pk)
	}

	tx.Sigs = append(tx.Sigs, types.Sig{
		SigData: make([][]byte, 0),
		M:       uint16(len(pubKeys) - (len(pubKeys)-1)/3),
		PubKeys: pubKeys,
	})
	sink := common.NewZeroCopySink(nil)
	if err := tx.Serialization(sink); err != nil {
		return err
	}

	fmt.Printf("raw transaction is %s\nNeed to send this transaction to every single consensus peer to sign. \n",
		hex.EncodeToString(sink.Bytes()))
	return nil
}

func CreateSyncMscGenesisHdrToPolyTx(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}

	h, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}

	ethRpc, err := cmd.Flags().GetString(MscRpcAddr)
	if err != nil {
		return err
	}
	et := NewEthTools(ethRpc)

	hdr, err := et.GetBlockHeader(h)
	if err != nil {
		return err
	}

	raw, err := json.Marshal(hdr)
	if err != nil {
		return err
	}

	poly := poly_go_sdk.NewPolySdk()
	tx, err := poly.Native.Hs.NewSyncGenesisHeaderTransaction(id, raw)
	if err != nil {
		return err
	}

	str, err := cmd.Flags().GetString(ConsensusPubKeys)
	pks := strings.Split(str, ",")
	if err != nil {
		return err
	}
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for i, v := range pks {
		pk, err := vconfig.Pubkey(v)
		if err != nil {
			return fmt.Errorf("failed to get no%d pubkey: %v", i, err)
		}
		pubKeys = append(pubKeys, pk)
	}

	tx.Sigs = append(tx.Sigs, types.Sig{
		SigData: make([][]byte, 0),
		M:       uint16(len(pubKeys) - (len(pubKeys)-1)/3),
		PubKeys: pubKeys,
	})
	sink := common.NewZeroCopySink(nil)
	if err := tx.Serialization(sink); err != nil {
		return err
	}

	fmt.Printf("raw transaction is %s\nNeed to send this transaction to every single consensus peer to sign. \n",
		hex.EncodeToString(sink.Bytes()))
	return nil
}

func CreateSyncBscGenesisHdrToPolyTx(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	h, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}

	ethRpc, err := cmd.Flags().GetString(BscRpcAddr)
	if err != nil {
		return err
	}
	et := NewEthTools(ethRpc)
	hdr, err := et.GetBlockHeader(h)
	if err != nil {
		return err
	}
	phdr, err := et.GetBlockHeader(h - 200)
	if err != nil {
		return err
	}
	pvalidators, err := bsc.ParseValidators(phdr.Extra[32 : len(phdr.Extra)-65])
	if err != nil {
		return err
	}

	if len(hdr.Extra) <= 65+32 {
		return fmt.Errorf("invalid epoch header at height: %d", h)
	}
	if len(phdr.Extra) <= 65+32 {
		return fmt.Errorf("invalid epoch header at height: %d", h-200)
	}

	genesisHeader := bsc.GenesisHeader{Header: *hdr, PrevValidators: []bsc.HeightAndValidators{
		{Height: big.NewInt(int64(h - 200)), Validators: pvalidators},
	}}
	raw, err := json.Marshal(genesisHeader)
	if err != nil {
		return err
	}

	poly := poly_go_sdk.NewPolySdk()
	tx, err := poly.Native.Hs.NewSyncGenesisHeaderTransaction(id, raw)
	if err != nil {
		return err
	}

	str, err := cmd.Flags().GetString(ConsensusPubKeys)
	pks := strings.Split(str, ",")
	if err != nil {
		return err
	}
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for i, v := range pks {
		pk, err := vconfig.Pubkey(v)
		if err != nil {
			return fmt.Errorf("failed to get no%d pubkey: %v", i, err)
		}
		pubKeys = append(pubKeys, pk)
	}

	tx.Sigs = append(tx.Sigs, types.Sig{
		SigData: make([][]byte, 0),
		M:       uint16(len(pubKeys) - (len(pubKeys)-1)/3),
		PubKeys: pubKeys,
	})
	sink := common.NewZeroCopySink(nil)
	if err := tx.Serialization(sink); err != nil {
		return err
	}

	fmt.Printf("raw transaction is %s\nNeed to send this transaction to every single consensus peer to sign. \n",
		hex.EncodeToString(sink.Bytes()))
	return nil
}

func CreateSyncRawGenesisHdrTxToPolyTx(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}

	rawHex, err := ioutil.ReadFile(args[1])
	if err != nil {
		return err
	}
	raw, err := hex.DecodeString(strings.TrimSpace(string(rawHex)))
	if err != nil {
		return err
	}

	poly := poly_go_sdk.NewPolySdk()
	tx, err := poly.Native.Hs.NewSyncGenesisHeaderTransaction(id, raw)
	if err != nil {
		return err
	}

	str, err := cmd.Flags().GetString(ConsensusPubKeys)
	pks := strings.Split(str, ",")
	if err != nil {
		return err
	}
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for i, v := range pks {
		pk, err := vconfig.Pubkey(v)
		if err != nil {
			return fmt.Errorf("failed to get no%d pubkey: %v", i, err)
		}
		pubKeys = append(pubKeys, pk)
	}

	tx.Sigs = append(tx.Sigs, types.Sig{
		SigData: make([][]byte, 0),
		M:       uint16(len(pubKeys) - (len(pubKeys)-1)/3),
		PubKeys: pubKeys,
	})
	sink := common.NewZeroCopySink(nil)
	if err := tx.Serialization(sink); err != nil {
		return err
	}

	fmt.Printf("raw transaction is %s\nNeed to send this transaction to every single consensus peer to sign. \n",
		hex.EncodeToString(sink.Bytes()))
	return nil
}

func CreateSyncOkGenesisHdrToPolyTx(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}

	rawHex, err := ioutil.ReadFile("raw.hex")
	if err != nil {
		return err
	}
	raw, err := hex.DecodeString(string(rawHex))
	if err != nil {
		return err
	}

	poly := poly_go_sdk.NewPolySdk()
	tx, err := poly.Native.Hs.NewSyncGenesisHeaderTransaction(id, raw)
	if err != nil {
		return err
	}

	str, err := cmd.Flags().GetString(ConsensusPubKeys)
	pks := strings.Split(str, ",")
	if err != nil {
		return err
	}
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for i, v := range pks {
		pk, err := vconfig.Pubkey(v)
		if err != nil {
			return fmt.Errorf("failed to get no%d pubkey: %v", i, err)
		}
		pubKeys = append(pubKeys, pk)
	}

	tx.Sigs = append(tx.Sigs, types.Sig{
		SigData: make([][]byte, 0),
		M:       uint16(len(pubKeys) - (len(pubKeys)-1)/3),
		PubKeys: pubKeys,
	})
	sink := common.NewZeroCopySink(nil)
	if err := tx.Serialization(sink); err != nil {
		return err
	}

	fmt.Printf("raw transaction is %s\nNeed to send this transaction to every single consensus peer to sign. \n",
		hex.EncodeToString(sink.Bytes()))
	return nil
}

func CreateSyncSwthGenesisHdrToPolyTx(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	h, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	swthRpc, err := cmd.Flags().GetString(SwitcheoRpcAddr)
	if err != nil {
		return err
	}
	rpcCli, err := http.New(swthRpc, "/websocket")
	if err != nil {
		return err
	}
	height := int64(h)
	res, err := rpcCli.Commit(&height)
	if err != nil {
		return err
	}
	vals, err := getValidators(rpcCli, height)
	if err != nil {
		panic(err)
	}
	ch := &cosmos.CosmosHeader{
		Header:  *res.Header,
		Commit:  res.Commit,
		Valsets: vals,
	}
	cdc := NewCodec()
	raw, err := cdc.MarshalBinaryBare(ch)
	if err != nil {
		return err
	}

	poly := poly_go_sdk.NewPolySdk()
	tx, err := poly.Native.Hs.NewSyncGenesisHeaderTransaction(id, raw)
	if err != nil {
		return err
	}

	str, err := cmd.Flags().GetString(ConsensusPubKeys)
	pks := strings.Split(str, ",")
	if err != nil {
		return err
	}
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for i, v := range pks {
		pk, err := vconfig.Pubkey(v)
		if err != nil {
			return fmt.Errorf("failed to get no%d pubkey: %v", i, err)
		}
		pubKeys = append(pubKeys, pk)
	}

	tx.Sigs = append(tx.Sigs, types.Sig{
		SigData: make([][]byte, 0),
		M:       uint16(len(pubKeys) - (len(pubKeys)-1)/3),
		PubKeys: pubKeys,
	})
	sink := common.NewZeroCopySink(nil)
	if err := tx.Serialization(sink); err != nil {
		return err
	}

	fmt.Printf("raw transaction is %s\nNeed to send this transaction to every single consensus peer to sign. \n",
		hex.EncodeToString(sink.Bytes()))
	return nil
}

func CreateSyncNeoGenesisHdrTx(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	h, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	rpcAddr, err := cmd.Flags().GetString(NeoRpcAddr)
	if err != nil {
		return err
	}
	cli := rpc.NewClient(rpcAddr)
	resp := cli.GetBlockHeaderByIndex(uint32(h))
	if resp.HasError() {
		return fmt.Errorf("failed to get header: %v", resp.Error.Message)
	}
	header, err := block.NewBlockHeaderFromRPC(&resp.Result)
	if err != nil {
		return err
	}
	buf := io.NewBufBinaryWriter()
	header.Serialize(buf.BinaryWriter)
	if buf.Err != nil {
		return buf.Err
	}

	poly := poly_go_sdk.NewPolySdk()
	tx, err := poly.Native.Hs.NewSyncGenesisHeaderTransaction(id, buf.Bytes())
	if err != nil {
		return err
	}

	str, err := cmd.Flags().GetString(ConsensusPubKeys)
	pks := strings.Split(str, ",")
	if err != nil {
		return err
	}
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for i, v := range pks {
		pk, err := vconfig.Pubkey(v)
		if err != nil {
			return fmt.Errorf("failed to get no%d pubkey: %v", i, err)
		}
		pubKeys = append(pubKeys, pk)
	}

	tx.Sigs = append(tx.Sigs, types.Sig{
		SigData: make([][]byte, 0),
		M:       uint16(len(pubKeys) - (len(pubKeys)-1)/3),
		PubKeys: pubKeys,
	})
	sink := common.NewZeroCopySink(nil)
	if err := tx.Serialization(sink); err != nil {
		return err
	}

	fmt.Printf("raw transaction is %s\nNeed to send this transaction to every single consensus peer to sign. \n",
		hex.EncodeToString(sink.Bytes()))
	return nil
}

func SignPolyMultiSigTx(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]

	tx := &types.Transaction{}
	raw, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}
	if err := tx.Deserialization(common.NewZeroCopySource(raw)); err != nil {
		return err
	}

	if err = poly.MultiSignToTransaction(tx, tx.Sigs[0].M, tx.Sigs[0].PubKeys, acc); err != nil {
		return fmt.Errorf("multi sign failed, err: %s", err)
	}

	sink := common.NewZeroCopySink(nil)
	err = tx.Serialization(sink)
	if err != nil {
		return err
	}

	if uint16(len(tx.Sigs[0].SigData)) >= tx.Sigs[0].M {
		txhash, err := poly.SendTransaction(tx)
		if err != nil {
			return fmt.Errorf("failed to send tx to poly: %v\nRaw Tx: %s\n", err, hex.EncodeToString(sink.Bytes()))
		}
		WaitPolyTx(txhash, poly)
		fmt.Printf("successful to sign poly tx and send tx %s to Poly with enough sigs\n", txhash.ToHexString())
		return nil
	}
	fmt.Printf("successful to sign tx and %d/%d sigs now, need at least %d sig: raw tx: %s\n", len(tx.Sigs[0].SigData),
		len(tx.Sigs[0].PubKeys), tx.Sigs[0].M, hex.EncodeToString(sink.Bytes()))

	return nil
}

func SyncPolyHdrToSwitcheo(cmd *cobra.Command, args []string) error {
	swthRpc, err := cmd.Flags().GetString(SwitcheoRpcAddr)
	if err != nil {
		return err
	}
	swthWallet, err := cmd.Flags().GetString(SwitcheoWallet)
	if err != nil {
		return err
	}
	swthPwd, err := cmd.Flags().GetString(SwitcheoWalletPwd)
	if err != nil {
		return err
	}
	if swthPwd == "" {
		fmt.Println("Pleasae input your switcheo wallet password...")
		pwd, err := password.GetPassword()
		if err != nil {
			return fmt.Errorf("getPassword error: %v", err)
		}
		swthPwd = string(pwd)
	}
	polyRpc, err := cmd.Flags().GetString(PolyRpcAddr)
	if err != nil {
		return err
	}
	h, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	gas, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	gasPrice, err := types2.ParseDecCoins(args[2])
	if err != nil {
		return err
	}
	fees, err := CalcCosmosFees(gasPrice, gas)
	if err != nil {
		return err
	}

	cli, err := http.New(swthRpc, "/websocket")
	if err != nil {
		return err
	}

	poly := poly_go_sdk.NewPolySdk()
	poly.NewRpcClient().SetAddress(polyRpc)
	hdr, err := poly.GetHeaderByHeight(uint32(h))
	if err != nil {
		return err
	}
	cdc := NewCodec()
	acc, err := NewCosmosAcc(swthWallet, swthPwd, cli, cdc)
	if err != nil {
		return err
	}

	res, seq, err := SendCosmosTx([]types2.Msg{&headersync.MsgSyncGenesisParam{
		Syncer:        acc.Acc,
		GenesisHeader: hex.EncodeToString(hdr.ToArray()),
	}}, acc, gas, fees, cdc, cli)
	if err != nil {
		return err
	}
	WaitSwitcheoTx(res.Hash, cli)

	hash := hdr.Hash()
	fmt.Printf("successful to sync poly header (hash: %s, height: %d) to Switcheo: (swth_txhash: %s, acc_seq: %d)\n",
		hash.ToHexString(), hdr.Height, res.Hash.String(), seq)

	return nil
}

func RegisterStateValidator(cmd *cobra.Command, args []string) error {
	stateValidatorString := args[0]
	svs := strings.Split(stateValidatorString, ",")

	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]
	txhash, err := poly.Native.Sm.RegisterStateValidator(svs, acc)
	if err != nil {
		return err
	}
	WaitPolyTx(txhash, poly)
	event, err := poly.GetSmartContractEvent(txhash.ToHexString())
	if err != nil {
		return err
	}
	var id uint64
	for _, e := range event.Notify {
		states := e.States.([]interface{})
		if states[0].(string) == "putStateValidatorApply" {
			id = uint64(states[1].(float64))
		}
	}
	fmt.Printf("successful to register state validators: %v, and id is: %d, txhash: %s\n", args, id, txhash.ToHexString())

	return nil
}

func ApproveRegisterStateValidator(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	wg.Add(len(accs))
	for _, acc := range accs {
		go func(acc *poly_go_sdk.Account) {
			defer wg.Done()
			txhash, err := poly.Native.Sm.ApproveRegisterStateValidator(id, acc)
			if err != nil {
				fmt.Printf("failed to approve state validators registration with acc: %s, %v\n", acc.Address.ToBase58(), err)
				return
			}
			WaitPolyTx(txhash, poly)
			fmt.Printf("successful to approve state validators registration id: %d, txhash: %s\n", id, txhash.ToHexString())
		}(acc)
	}
	wg.Wait()

	return nil
}

func RemoveStateValidator(cmd *cobra.Command, args []string) error {
	stateValidatorString := args[0]
	svs := strings.Split(stateValidatorString, ",")

	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	acc := accs[0]
	txhash, err := poly.Native.Sm.RemoveStateValidator(svs, acc)
	if err != nil {
		return err
	}
	WaitPolyTx(txhash, poly)
	event, err := poly.GetSmartContractEvent(txhash.ToHexString())
	if err != nil {
		return err
	}
	var id uint64
	for _, e := range event.Notify {
		states := e.States.([]interface{})
		if states[0].(string) == "putStateValidatorRemove" {
			id = uint64(states[1].(float64))
		}
	}
	fmt.Printf("successful to remove state validators: %v, and id is: %d, txhash: %s\n", args, id, txhash.ToHexString())

	return nil
}

func ApproveRemoveStateValidator(cmd *cobra.Command, args []string) error {
	poly, accs, err := GetPolyAndAccsByCmd(cmd)
	if err != nil {
		return err
	}
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	wg.Add(len(accs))
	for _, acc := range accs {
		go func(acc *poly_go_sdk.Account) {
			defer wg.Done()
			txhash, err := poly.Native.Sm.ApproveRemoveStateValidator(id, acc)
			if err != nil {
				fmt.Printf("failed to approve state validators removal with acc: %s, %v\n", acc.Address.ToBase58(), err)
				return
			}
			WaitPolyTx(txhash, poly)
			fmt.Printf("successful to approve state validators removal id: %d, txhash: %s\n", id, txhash.ToHexString())
		}(acc)
	}
	wg.Wait()

	return nil
}
