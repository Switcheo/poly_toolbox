<h1 align="center">Toolbox For Poly</h1>

## Introduction

Toolbox is a helper for governance of poly network. To Poly consensus node, you can finish management work by toolbox. Like making a proposal to register a new consensus node for next round consensus, etc.

## Build From Source

### Prerequisites

- [Golang](https://golang.org/doc/install) version 1.14 or later

### Build

```shell
git clone https://github.com/polynetwork/poly_toolbox.git
cd poly_toolbox
go build -o toolbox cmd/run.go
```

After building the source code successfully,  you should see the executable program `toolbox`.

You can see some guide information by running `toolbox -h`.

## Usage

### Register Relayer

If you want to send transaciton to Poly, you need to be a relayer account first. Do as follow:

```
./toolbox poly relayer_manager register_relayer [address_to_register] --poly_rpc_addr="http://poly_rpc:port" --signer_wallet_path=/path/to/wallet.dat
```

The `signer` must be a relayer. You will  find the ID of this registration proposal in the result. Remember it will be useful during `approve` part.

You can set password by flag `--signer_wallet_pwd`. But we suggest you **not using** it. And toolbox will require you to input password when running.

### Approve Relayer Registration

Consensus account can approve registration of relayer with the registration **ID**.

```
./toolbox poly relayer_manager approve_register_relayer [ID] --poly_rpc_addr="http://poly_rpc:port" --signer_wallet_path=/path/to/wallet.dat
```

You can find the transaction hash in result.

### Register Side Chain

Register a side chain to Poly cross-chain system.

```
./toolbox poly side_chain_manager register_side_chain --chain_id=[number] --router=[number] --name=[name] --blocks_to_wait=1 --CMCC="hex" --poly_rpc_addr="http://poly_rpc:port" --signer_wallet_path=/path/to/wallet.dat
```

For those chains with `extra_info` field, we need to add one more flag `--extra`. For example.

```
./toolbox poly side_chain_manager register_side_chain --chain_id=[number] --router=[number] --name=[name] --blocks_to_wait=1 --extra='json' --CMCC="hex" --poly_rpc_addr="http://poly_rpc:port" --signer_wallet_path=/path/to/wallet.dat
```

The extra info should be a string in json format, like `{"ChainID":6}`.

### Approve Side Chain Registration

Consensus account can approve registration of a side chain with the chain-id.

```
./toolbox poly side_chain_manager approve_register_side_chain --chain_id=[number] --poly_rpc_addr="http://poly_rpc:port" --signer_wallet_path=/path/to/wallet.dat
```

### Create Transaction To Sync Ontology Genesis Header

Create a raw transaction to sync Ontology genesis header to Poly. Noted, you must set all public keys of multisig address.

```
./toolbox poly header_sync create_sync_ont_genesis_hdr_tx [chain-id] [height] --consensus_public_keys=pub1,pub2,pub3 --ont_rpc="http://ontology:20336"
```

### Sign Multisig Transacion of Poly

Using your wallet to sign a raw multisig-transaction. Return with the raw transacion signed by your account. Pass the raw transaction to next one of the multisig-address.

```
./toolbox poly header_sync sign_poly_multisig_tx [raw_tx] --poly_rpc_addr="http://poly_rpc_addr" --signer_wallet_path=/path/to/wallet.dat
```

### Create Transaction To Sync Switcheo Genesis Header

Create a poly transaction for syncing Switcheo genesis header to Poly. You need to set the cross-chain chain-id of Switcheo and the height of Switcheo header. This block header is not necessarily the genesis block header of Switcheo. "Genesis" means that the header is a genesis start for cross-chain business. So the header can be any one on Switcheo chain and the lastest one is suggested.

Run the follow and get the raw transaction:

```
./toolbox poly header_sync create_sync_switcheo_genesis_hdr_tx [carbon_chain_id] [carbon_hdr_height] --switcheo_rpc="http://switcheo_rpc" --consensus_public_keys=pub1,pub2,pub3
```

So you have the raw transaction of Poly and you need to sign this transaction using `./toolbox poly header_sync sign_poly_multisig_tx` which has already been introduced above. After signing, you need to send it to next signer in consensus peers. Signers are going to sign it one by one. When the signatures is enough, the transaction would be send to Poly automaticly.

### Create Transaction To Sync NEO Genesis Header

Create a poly transaction for syncing NEO genesis header to Poly. You need to set the cross-chain chain-id of NEO and the height of NEO header.

Run the follow and get the raw transaction:

```
./toolbox poly header_sync create_sync_neo_genesis_hdr_tx [neo_chain_id] [neo_hdr_height] --neo_rpc="http://neo_rpc" --consensus_public_keys=pub1,pub2,pub3
```

Next sign it by `./toolbox poly header_sync sign_poly_multisig_tx`.

### Create Transaction To Sync Ethereum Genesis Header

Create a poly transaction for syncing Ethereum genesis header to Poly. You need to set the cross-chain chain-id of Ethereum and the height of Ethereum header.

Run the follow and get the raw transaction:

```
./toolbox poly header_sync create_sync_eth_genesis_hdr_tx [eth_chain_id] [eth_hdr_height] --eth_rpc="http://eth_rpc" --consensus_public_keys=pub1,pub2,pub3
```

Next sign it by `./toolbox poly header_sync sign_poly_multisig_tx`.

### Create Transaction To Sync BSC Genesis Header

Create a poly transaction for syncing BSC genesis header to Poly. You need to set the cross-chain chain-id of BSC and the height of BSC header.

Run the follow and get the raw transaction:

```
./toolbox poly header_sync create_sync_bsc_genesis_hdr_tx [bsc_chain_id] [bsc_epoch_hdr_height] --bsc_rpc="http://bsc_rpc" --consensus_public_keys=pub1,pub2,pub3
```

Next sign it by `./toolbox poly header_sync sign_poly_multisig_tx`.

### Sync Poly Genesis Header to Switcheo

When switcheo init poly cross-chain function, you need to sync poly header to Switcheo. This header must be consensus switching block header. That means field `ConsensusPayload` of header is not empty. The lastest switching header is suggested.

```
./toolbox switcheo sync_poly_genesis_hdr_to_switcheo [height] [carbon_gas] [carbon_price] --poly_rpc_addr="http://poly_rpc_addr" --switcheo_rpc="http://switcheo_rpc" --switcheo_wallet=/path/to/wallet
```

The `switcheo_wallet` should be exported from tendermint. Like follow:

```
-----BEGIN TENDERMINT PRIVATE KEY-----
kdf: bcrypt
salt: 9F6C572C2ECBCCE0A8785D22069AE2E2
type: secp256k1

hS+xd9VuT2svdqC2g8f0abcd72PS6Wld+GQHWNE/QqDVbcHJO/X4TsqXHBfV55+h
CBfy121ENrs80OigsnlZEzU+fWo0OhW1V1O2j1M=
=l8jB
-----END TENDERMINT PRIVATE KEY-----
```