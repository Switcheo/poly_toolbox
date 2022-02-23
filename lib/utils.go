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
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
	types3 "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/staking"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/polynetwork/cosmos-poly-module/btcx"
	"github.com/polynetwork/cosmos-poly-module/ccm"
	"github.com/polynetwork/cosmos-poly-module/ft"
	"github.com/polynetwork/cosmos-poly-module/headersync"
	"github.com/polynetwork/cosmos-poly-module/lockproxy"
	poly_go_sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/password"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/crypto"
	bytes2 "github.com/tendermint/tendermint/libs/bytes"
	http2 "github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	types2 "github.com/tendermint/tendermint/types"
)

const (
	SignerWalletPath     = "signer_wallet_path"
	SignerWalletPwd      = "signer_wallet_pwd"
	PolyRpcAddr          = "poly_rpc_addr"
	ConsensusPubKeys     = "consensus_public_keys"
	ChainId              = "chain_id"
	Router               = "router"
	Name                 = "name"
	BlkToWait            = "blocks_to_wait"
	CMCC                 = "CMCC"
	ExtraInfo            = "extra"
	OntRpcAddr           = "ont_rpc"
	EthRpcAddr           = "eth_rpc"
	BscRpcAddr           = "bsc_rpc"
	MscRpcAddr           = "msc_rpc"
	CarbonRpcAddr        = "carbon_rpc"
	NeoRpcAddr           = "neo_rpc"
	CarbonWallet         = "carbon_wallet"
	CarbonWalletPwd      = "carbon_wallet_pwd"
	CarbonCosmosChainID  = "carbon-1"
	BroadcastConnTimeOut = "connection timed out"
	SeqErr               = "verify correct account sequence and chain-id"
)

func GetPolyAccountByPassword(asdk *poly_go_sdk.PolySdk, path, pwdStr string) (*poly_go_sdk.Account, error) {
	wallet, err := asdk.OpenWallet(path)
	if err != nil {
		return nil, fmt.Errorf("open wallet error: %v", err)
	}
	var pwd []byte
	if pwdStr != "" {
		pwd = []byte(pwdStr)
	} else {
		fmt.Println("Pleasae input your poly wallet password...")
		pwd, err = password.GetPassword()
		if err != nil {
			return nil, fmt.Errorf("getPassword error: %v", err)
		}
	}
	user, err := wallet.GetDefaultAccount(pwd)
	if err != nil {
		return nil, fmt.Errorf("getDefaultAccount error: %v", err)
	}
	return user, nil
}

func SetUpPoly(poly *poly_go_sdk.PolySdk, rpc string) error {
	poly.NewRpcClient().SetAddress(rpc)
	hdr, err := poly.GetHeaderByHeight(0)
	if err != nil {
		return err
	}
	poly.SetChainId(hdr.ChainID)
	return nil
}

func WaitPolyTx(txhash common.Uint256, poly *poly_go_sdk.PolySdk) {
	fmt.Printf("waiting poly transaction %s confirmed...\n", txhash.ToHexString())
	tick := time.NewTicker(100 * time.Millisecond)
	var h uint32
	startTime := time.Now()
	for range tick.C {
		h, _ = poly.GetBlockHeightByTxHash(txhash.ToHexString())
		curr, _ := poly.GetCurrentBlockHeight()
		if h > 0 && curr > h {
			break
		}

		if startTime.Add(100 * time.Millisecond); startTime.Second() > 300 {
			panic(fmt.Errorf("tx( %s ) is not confirm for a long time ( over %d sec )",
				txhash.ToHexString(), 300))
		}
	}
}

func GetPolyAndAccsByCmd(cmd *cobra.Command) (*poly_go_sdk.PolySdk, []*poly_go_sdk.Account, error) {
	poly := poly_go_sdk.NewPolySdk()
	rpcAddr, err := cmd.Flags().GetString(PolyRpcAddr)
	if err != nil {
		return nil, nil, err
	}
	walletPath, err := cmd.Flags().GetString(SignerWalletPath)
	if err := SetUpPoly(poly, rpcAddr); err != nil {
		return nil, nil, err
	}
	pwd, err := cmd.Flags().GetString(SignerWalletPwd)
	if err != nil {
		return nil, nil, err
	}
	wallet, err := poly.OpenWallet(walletPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open wallet error: %v", err)
	}
	i, num := 0, wallet.GetAccountCount()
	if num == 0 {
		return nil, nil, fmt.Errorf("No wallets found in %s", walletPath)
	}
	accs := make([]*poly_go_sdk.Account, num)
	fmt.Printf("Found %d wallets in %s\n", num, walletPath)
	for i < num {
		accs[i], err = wallet.GetAccountByIndex(i+1, []byte(pwd))
		if err != nil {
			return nil, nil, fmt.Errorf("GetAccountByIndex error: %v", err)
		}
		i++
	}
	return poly, accs, nil
}

type ETHTools struct {
	restclient *RestClient
	ethclient  *ethclient.Client
}

type LockEvent struct {
	Method   string
	TxHash   string
	Txid     []byte
	Saddress string
	Tchain   uint32
	Taddress string
	Height   uint64
	Value    []byte
}
type UnlockEvent struct {
	Method   string
	Txid     string
	RTxid    string
	FromTxId string
	Height   uint64
	Token    string
}

type heightReq struct {
	JsonRpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      uint     `json:"id"`
}

type heightRep struct {
	JsonRpc string `json:"jsonrpc"`
	Result  string `json:"result"`
	Id      uint   `json:"id"`
}

type BlockReq struct {
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      uint          `json:"id"`
}

type BlockRep struct {
	JsonRPC string        `json:"jsonrpc"`
	Result  *types.Header `json:"result"`
	Id      uint          `json:"id"`
}

func NewEthTools(url string) *ETHTools {
	ethclient, err := ethclient.Dial(url)
	if err != nil {
		return nil
	}
	restclient := NewRestClient()
	restclient.SetAddr(url)
	tool := &ETHTools{
		restclient: restclient,
		ethclient:  ethclient,
	}
	return tool
}

func (self *ETHTools) GetEthClient() *ethclient.Client {
	return self.ethclient
}

func (self *ETHTools) GetNodeHeight() (uint64, error) {
	req := &heightReq{
		JsonRpc: "2.0",
		Method:  "eth_blockNumber",
		Params:  make([]string, 0),
		Id:      1,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return 0, fmt.Errorf("GetNodeHeight: marshal req err: %s", err)
	}
	resp, err := self.restclient.SendRestRequest(data)
	if err != nil {
		return 0, fmt.Errorf("GetNodeHeight err: %s", err)
	}
	rep := &heightRep{}
	err = json.Unmarshal(resp, rep)
	if err != nil {
		return 0, fmt.Errorf("GetNodeHeight, unmarshal resp err: %s", err)
	}
	height, err := strconv.ParseUint(rep.Result, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("GetNodeHeight, parse resp height %s failed", rep.Result)
	} else {
		return height, nil
	}
}

func (self *ETHTools) GetBlockHeader(height uint64) (*types.Header, error) {
	params := []interface{}{fmt.Sprintf("0x%x", height), true}
	req := &BlockReq{
		JsonRpc: "2.0",
		Method:  "eth_getBlockByNumber",
		Params:  params,
		Id:      1,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("GetNodeHeight: marshal req err: %s", err)
	}
	resp, err := self.restclient.SendRestRequest(data)
	if err != nil {
		return nil, fmt.Errorf("GetNodeHeight err: %s", err)
	}
	rsp := &BlockRep{}
	err = json.Unmarshal(resp, rsp)
	if err != nil {
		return nil, fmt.Errorf("GetNodeHeight, unmarshal resp err: %s", err)
	}

	return rsp.Result, nil
}

func (self *ETHTools) WaitTransactionsConfirm(hashs []common2.Hash) {
	hasPending := true
	for hasPending {
		time.Sleep(time.Second * 1)
		hasPending = false
		for _, hash := range hashs {
			_, ispending, err := self.ethclient.TransactionByHash(context.Background(), hash)
			if err != nil {
				hasPending = true
				continue
			}
			if ispending == true {
				hasPending = true
			} else {
			}
		}
	}
}

func (self *ETHTools) WaitTransactionConfirm(hash common2.Hash) {
	for {
		time.Sleep(time.Second * 1)
		_, ispending, err := self.ethclient.TransactionByHash(context.Background(), hash)
		if err != nil {
			continue
		}
		if ispending == true {
			continue
		} else {
			break
		}
	}

}

type RestClient struct {
	addr       string
	restClient *http.Client
	user       string
	passwd     string
}

func NewRestClient() *RestClient {
	return &RestClient{
		restClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   5,
				DisableKeepAlives:     false,
				IdleConnTimeout:       time.Second * 300,
				ResponseHeaderTimeout: time.Second * 300,
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: time.Second * 300,
		},
	}
}

func (self *RestClient) SetAddr(addr string) *RestClient {
	self.addr = addr
	return self
}

func (self *RestClient) SetAuth(user string, passwd string) *RestClient {
	self.user = user
	self.passwd = passwd
	return self
}

func (self *RestClient) SetRestClient(restClient *http.Client) *RestClient {
	self.restClient = restClient
	return self
}

func (self *RestClient) SendRestRequest(data []byte) ([]byte, error) {
	resp, err := self.restClient.Post(self.addr, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("http post request:%s error:%s", data, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rest response body error:%s", err)
	}
	return body, nil
}

func (self *RestClient) SendRestRequestWithAuth(data []byte) ([]byte, error) {
	url := self.addr
	bodyReader := bytes.NewReader(data)
	httpReq, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("SendRestRequestWithAuth - build http request error:%s", err)
	}
	httpReq.Close = true
	httpReq.Header.Set("Content-Type", "application/json")

	httpReq.SetBasicAuth(self.user, self.passwd)

	rsp, err := self.restClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("SendRestRequestWithAuth - http post error:%s", err)
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil || len(body) == 0 {
		return nil, fmt.Errorf("SendRestRequestWithAuth - read rest response body error:%s", err)
	}
	return body, nil
}

func getValidators(rpc *http2.HTTP, h int64) ([]*types2.Validator, error) {
	p := 1
	vSet := make([]*types2.Validator, 0)
	for {
		res, err := rpc.Validators(&h, p, 100)
		if err != nil {
			if strings.Contains(err.Error(), "page should be within") {
				return vSet, nil
			}
			return nil, err
		}
		// In case tendermint don't give relayer the right error
		if len(res.Validators) == 0 {
			return vSet, nil
		}
		vSet = append(vSet, res.Validators...)
		p++
	}
}

func NewCodec() *codec.Codec {
	cdc := codec.New()
	bank.RegisterCodec(cdc)
	types3.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	auth.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	btcx.RegisterCodec(cdc)
	ccm.RegisterCodec(cdc)
	ft.RegisterCodec(cdc)
	headersync.RegisterCodec(cdc)
	lockproxy.RegisterCodec(cdc)
	return cdc
}

type CosmosAcc struct {
	Acc        types3.AccAddress
	PrivateKey crypto.PrivKey
	Seq        *CosmosSeq
	AccNum     uint64
}

func NewCosmosAcc(wallet, pwd string, cli *http2.HTTP, cdc *codec.Codec) (*CosmosAcc, error) {
	config := types3.GetConfig()
	config.SetBech32PrefixForAccount("carbon", "carbonpub")
	config.SetBech32PrefixForValidator("carbonvaloper", "carbonvaloperpub")
	config.SetBech32PrefixForConsensusNode("carbonvalcons", "carbonvalconspub")

	acc := &CosmosAcc{}
	bz, err := ioutil.ReadFile(wallet)
	if err != nil {
		return nil, err
	}

	privKey, _, err := mintkey.UnarmorDecryptPrivKey(string(bz), string(pwd))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %v", err)
	}

	acc.PrivateKey = privKey
	acc.Acc = types3.AccAddress(privKey.PubKey().Address().Bytes())
	var eAcc exported.Account
	rawParam, err := cdc.MarshalJSON(auth.NewQueryAccountParams(acc.Acc))
	if err != nil {
		return nil, err
	}
	res, err := cli.ABCIQuery("/custom/acc/account", rawParam)
	if err != nil {
		return nil, err
	}
	if !res.Response.IsOK() {
		return nil, fmt.Errorf("failed to get response for accout-query: %v", res.Response)
	}
	if err := cdc.UnmarshalJSON(res.Response.Value, &eAcc); err != nil {
		return nil, fmt.Errorf("unmarshal query-account-resp failed, err: %v", err)
	}
	acc.Seq = &CosmosSeq{
		lock: sync.Mutex{},
		val:  eAcc.GetSequence(),
	}
	acc.AccNum = eAcc.GetAccountNumber()

	return acc, nil
}

type CosmosSeq struct {
	lock sync.Mutex
	val  uint64
}

func (seq *CosmosSeq) GetAndAdd() uint64 {
	seq.lock.Lock()
	defer func() {
		seq.val += 1
		seq.lock.Unlock()
	}()
	return seq.val
}

func SendCosmosTx(msgs []types3.Msg, acc *CosmosAcc, gas uint64, fees types3.Coins, cdc *codec.Codec, cli *http2.HTTP) (*coretypes.ResultBroadcastTx, uint64, error) {
	seq := acc.Seq.GetAndAdd()
	toSign := auth.StdSignMsg{
		Sequence:      seq,
		AccountNumber: acc.AccNum,
		ChainID:       CarbonCosmosChainID,
		Msgs:          msgs,
		Fee:           auth.NewStdFee(gas, fees),
	}
	sig, err := acc.PrivateKey.Sign(toSign.Bytes())
	if err != nil {
		return nil, seq, fmt.Errorf("failed to sign raw tx: (error: %v, raw tx: %x)", err, toSign.Bytes())
	}

	tx := auth.NewStdTx(msgs, toSign.Fee, []auth.StdSignature{{acc.PrivateKey.PubKey(), sig}}, toSign.Memo)
	encoder := auth.DefaultTxEncoder(cdc)
	rawTx, err := encoder(tx)
	if err != nil {
		return nil, seq, fmt.Errorf("failed to encode signed tx: %v", err)
	}
	var res *coretypes.ResultBroadcastTx
	for {
		res, err = cli.BroadcastTxSync(rawTx)
		if err != nil {
			if strings.Contains(err.Error(), BroadcastConnTimeOut) {
				time.Sleep(10 * time.Second)
				continue
			}
			return nil, seq, fmt.Errorf("failed to broadcast tx: (error: %v, raw tx: %x)", err, rawTx)
		}
		if res.Code != 0 {
			if strings.Contains(res.Log, SeqErr) {
				time.Sleep(time.Second)
				continue
			}
			return nil, seq, fmt.Errorf("failed to check tx: (code: %d, sequence: %d, log: %s)", res.Code, seq, res.Log)
		} else {
			break
		}
	}

	return res, seq, nil
}

func CalcCosmosFees(gasPrice types3.DecCoins, gas uint64) (types3.Coins, error) {
	if gasPrice.IsZero() {
		return types3.Coins{}, errors.New("gas price is zero")
	}
	if gas == 0 {
		return types3.Coins{}, errors.New("gas is zero")
	}
	glDec := types3.NewDec(int64(gas))
	fees := make(types3.Coins, len(gasPrice))
	for i, gp := range gasPrice {
		fee := gp.Amount.Mul(glDec)
		fees[i] = types3.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	}
	return fees, nil
}

func WaitCarbonTx(txhash bytes2.HexBytes, cli *http2.HTTP) {
	tick := time.NewTicker(time.Second)
	for range tick.C {
		res, err := cli.Tx(txhash, false)
		if err == nil && res.Height > 0 {
			break
		}
	}
}
