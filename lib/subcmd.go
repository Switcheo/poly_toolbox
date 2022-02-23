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
	"github.com/spf13/cobra"
)

func PolyCmd() *cobra.Command {
	nmCmd := &cobra.Command{
		Use:   "poly",
		Short: "Poly subcommands",
		Long: "By this command, you can register relayer and side chain. And you can introduce other account " +
			"to join the consensus peers. It's about governance of Poly.",
	}

	nmCmd.PersistentFlags().String(PolyRpcAddr, "", "set poly rpc address")
	nmCmd.PersistentFlags().String(SignerWalletPath, "", "The signer's wallet file to sign poly transaction "+
		"registering candidate. ")
	nmCmd.PersistentFlags().String(SignerWalletPwd, "", "The password for signer's wallet. "+
		"If not set here, you will be required to input it. ")

	nmCmd.AddCommand(PolyNMCmd(), PolyRMCmd(), PolySMCmd(), PolyHeaderSyncCmd(), PolyNSMCmd())

	return nmCmd
}

func PolyNMCmd() *cobra.Command {
	nmCmd := &cobra.Command{
		Use:   "node_manager",
		Short: "Poly node_manager subcommands",
		Long: "Node manager is part of poly governance which hold the information of consensus peers. " +
			"Including peers joining or quitting consensus and peers blacklist.",
	}

	nmCmd.AddCommand(
		RegisterCandidateCmd(),
		ApproveCandidateCmd(),
		UnRegisterCandidateCmd(),
		QuitNodeCmd(),
		BlackNodeCmd(),
		WhiteNodeCmd(),
		CreateCommitDposTxCmd(),
		SignCommitDposTxCmd(),
		CreateUpdateConfigTxCmd(),
		SignUpdateConfigTxCmd())

	return nmCmd
}

func RegisterCandidateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "register_candidate [candidate_public_key]",
		Short: "introduce a new account to consensus. ",
		Long: "Register a new candidate for node which want to be a new consensus node of Poly. " +
			"Only one argument candidate's public key which can be found in candidate's wallet. ",
		RunE: RegisterCandidate,
		Args: cobra.ExactArgs(1),
	}
	return c
}

func ApproveCandidateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "approve_candidate [candidate_public_key]",
		Short: "approve candidate's registration",
		Long:  "Using consensus peer wallet to approve candidate's registration. ",
		RunE:  ApproveCandidate,
		Args:  cobra.ExactArgs(1),
	}
	return c
}

func UnRegisterCandidateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "unregister_candidate",
		Short: "cancel your relayer registration",
		Long:  "Unregister a candidate and only candidate can unregister itself. ",
		RunE:  UnRegisterCandidate,
	}
	return c
}

func QuitNodeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "quit_node",
		Short: "leave the consensus peer set",
		Long:  "Quit from consensus peer pool of Poly. Signer must be quit-node itself. ",
		RunE:  QuitNode,
	}
	return c
}

func BlackNodeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "black_node [node_public_key]",
		Short: "pull a node into the blacklist",
		Long: "Vote for putting a node into black list on Poly. If over 2/3 consensus peers vote yes, that node will " +
			"be blacked. It can't be consensus util it get moved from list. Signer must be a consensus account",
		RunE: BlackNode,
		Args: cobra.ExactArgs(1),
	}
	return c
}

func WhiteNodeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "white_node [node_public_key]",
		Short: "pull a node out of the blacklist",
		Long: "Vote for moving node out of black list on Poly. If over 2/3 consensus peers vote yes, that node will " +
			"be whited. Signer must be a consensus account",
		RunE: WhiteNode,
		Args: cobra.ExactArgs(1),
	}
	return c
}

func CreateCommitDposTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "create_raw_commit_dpos_tx",
		Short: "create transaction that actively switch the consensus cycle",
		Long: "Create a unsigned transaction for committing DPOS. Poly will step into next epoch of consensus if " +
			"this transacion execute. ",
		RunE: CreateCommitDposTx,
	}
	c.Flags().String(ConsensusPubKeys, "", "public keys for consensus peers, sep by ','. ")
	return c
}

func SignCommitDposTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "sign_commit_dpos_tx [raw_tx]",
		Short: "sign transaction that actively switch the consensus cycle",
		Long: "Using your wallet to sign a raw transaction and return the signed raw transaction. It will tell you " +
			"if transaction is good to send. ",
		RunE: SignCommitDposTx,
		Args: cobra.ExactArgs(1),
	}
	return c
}

func CreateUpdateConfigTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "create_update_config_tx [blockMsgDelay] [hashMsgDelay] [peerHandshakeTimeout] [maxBlockChangeView]",
		Short: "create transaction that update poly governance configuration",
		Long: "Using your wallet to sign a raw transaction and return the signed raw transaction. It will tell you " +
			"if transaction is good to send. ",
		RunE: CreateUpdateConfigTx,
		Args: cobra.ExactArgs(4),
	}
	c.Flags().String(ConsensusPubKeys, "", "public keys for consensus peers, sep by ','. ")
	return c
}

func SignUpdateConfigTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "sign_update_config_tx [raw_tx]",
		Short: "sign transaction that update poly governance configuration",
		Long: "Using your wallet to sign a raw transaction and return the signed raw transaction. It will tell you " +
			"if transaction is good to send. ",
		RunE: SignUpdateConfigTx,
		Args: cobra.ExactArgs(1),
	}
	return c
}

func PolyRMCmd() *cobra.Command {
	rm := &cobra.Command{
		Use:   "relayer_manager",
		Short: "Poly relayer manager subcommands",
		Long:  "relayer manager control relayer accout set. Like becoming a relayer or be removed.",
	}
	rm.AddCommand(
		RegisterRelayerCmd(),
		ApproveRegisterRelayerCmd(),
		RemoveRelayerCmd(),
		ApproveRemoveRelayerCmd())
	return rm
}

func RegisterRelayerCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "register_relayer [address] [address] ...",
		Short: "register new addresses as poly relayer. ",
		Long: "You can propose registration for a dozen of addresses in one transaction and toolbox " +
			"will return the registration ID which is needed when approving. ",
		RunE: RegisterRelayer,
		Args: cobra.MinimumNArgs(1),
	}
	return c
}

func ApproveRegisterRelayerCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "approve_register_relayer [id]",
		Short: "approve registration of proposal ID",
		Long:  "Consensus account approve the registration with the ID returned when registering.",
		RunE:  ApproveRegisterRelayer,
		Args:  cobra.ExactArgs(1),
	}
	return c
}

func RemoveRelayerCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "remove_relayer [address] [address] ...",
		Short: "remove addresses as poly relayer. ",
		Long:  "Remove address from relayer set with returning the proposal ID",
		RunE:  RemoveRelayer,
	}
	return c
}

func ApproveRemoveRelayerCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "approve_remove_relayer [id]",
		Short: "approve removing proposal ID as poly relayer. ",
		Long:  "Consensus account approve the relayer removing with the ID returned when removing.",
		RunE:  ApproveRemoveRelayer,
	}
	return c
}

func PolySMCmd() *cobra.Command {
	sm := &cobra.Command{
		Use:   "side_chain_manager",
		Short: "Poly side chain manager subcommands",
		Long: "Poly side chain manager hold the information of chains alive in cross-chain system. " +
			"You can register a new chain or remove it. ",
	}
	sm.AddCommand(
		RegisterSideChainCmd(),
		ApproveRegisterSideChainCmd(),
		UpdateSideChainCmd(),
		ApproveUpdateSideChainCmd(),
		QuitSideChainCmd(),
		ApproveQuitSideChainCmd())
	return sm
}

func RegisterSideChainCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "register_side_chain",
		Short: "register a new chain",
		Long: "Sending transaction to register a new chain in cross-chain system. You need to " +
			"set parameters in flags including chain_id, router, name, blocks_to_wait and CMCC. ",
		RunE: RegisterSideChain,
	}

	c.Flags().Uint64(ChainId, 0, "set chain id for new chain")
	c.Flags().Uint64(Router, 0, "set chain router")
	c.Flags().String(Name, "", "set chain name")
	c.Flags().Uint64(BlkToWait, 1, "set block number to confirm a transacion on this chain")
	c.Flags().String(CMCC, "", "set cmcc address")
	c.Flags().String(ExtraInfo, "", "extra info for this chain in json")

	return c
}

func ApproveRegisterSideChainCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "approve_register_side_chain",
		Short: "approve registration of chain-id",
		Long:  "Consensus account approve the side chain registration with the chain-id.",
		RunE:  ApproveRegisterSideChain,
	}

	c.Flags().Uint64(ChainId, 0, "set chain id for new chain")

	return c
}

func UpdateSideChainCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "update_side_chain",
		Short: "update side chain information. ",
		Long: "Send transaction to make a proposal for updating chain configuration. Only " +
			"account registering this chain. You need to set parameters in flags including " +
			"chain_id, router, name, blocks_to_wait and CMCC.",
		RunE: UpdateSideChain,
	}

	c.Flags().Uint64(ChainId, 0, "set chain id for new chain")
	c.Flags().Uint64(Router, 0, "set chain router")
	c.Flags().String(Name, "", "set chain name")
	c.Flags().Uint64(BlkToWait, 1, "set block number to confirm a transacion on this chain")
	c.Flags().String(CMCC, "", "set cmcc address")
	c.Flags().String(ExtraInfo, "", "extra info for this chain in json")

	return c
}

func ApproveUpdateSideChainCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "approve_update_side_chain",
		Short: "approve update side chain information. ",
		Long: "Approve the proposal for updating the chain of chain-id. Only consensus accounts " +
			"can send transaction to approve.",
		RunE: ApproveUpdateSideChain,
	}

	c.Flags().Uint64(ChainId, 0, "set chain id to identify the chain")

	return c
}

func QuitSideChainCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "quit_side_chain",
		Short: "make the side chain quit cross-chain system.",
		Long: "Make a proposal to remove a chain from cross-chain system. Only account registering " +
			"this chain can do this. ",
		RunE: QuitSideChain,
	}

	c.Flags().Uint64(ChainId, 0, "set chain id to identify the chain")

	return c
}

func ApproveQuitSideChainCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "approve_quit_side_chain",
		Short: "approve quit side chain. only consensus accounts.",
		Long: "Approve the proposal that try to remove a side chain from cross-chain system. " +
			"Only consensus accounts can send transaction to approve.",
		RunE: ApproveQuitSideChain,
	}

	c.Flags().Uint64(ChainId, 0, "set chain id to identify the chain")

	return c
}

func PolyHeaderSyncCmd() *cobra.Command {
	sm := &cobra.Command{
		Use:   "header_sync",
		Short: "sync genesis header subcommands",
		Long:  "Sync genesis headers to Poly. After this, the chains registered on Poly are good to cross chain. ",
	}
	sm.AddCommand(
		CreateSyncOntGenesisHdrTxCmd(),
		CreateSyncEthGenesisHdrTxCmd(),
		CreateSyncSwticheoGenesisHdrTxCmd(),
		CreateSyncNeoGenesisHdrTxCmd(),
		CreateSyncBscGenesisHdrTxCmd(),
		CreateSyncMscGenesisHdrTxCmd(),
		CreateSyncOkGenesisHdrTxCmd(),
		CreateSyncRawGenesisHdrTxCmd(),
		SignPolyMultiSigTxCmd())
	return sm
}

func CreateSyncOntGenesisHdrTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "create_sync_ont_genesis_hdr_tx [ont_chain_id] [ont_hdr_height]",
		Short: "create transaction to sync ontology epoch switching header to Poly.",
		RunE:  CreateSyncOntGenesisHdrToPolyTx,
	}

	c.Flags().String(OntRpcAddr, "", "ontology node RPC address")
	c.Flags().String(ConsensusPubKeys, "", "public keys for consensus peers, sep by ','. ")
	_ = c.MarkFlagRequired(OntRpcAddr)

	return c
}

func CreateSyncEthGenesisHdrTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "create_sync_eth_genesis_hdr_tx [eth_chain_id] [eth_hdr_height]",
		Short: "create transaction to sync ethereum header to Poly.",
		RunE:  CreateSyncEthGenesisHdrToPolyTx,
	}

	c.Flags().String(EthRpcAddr, "", "ethereum node RPC address")
	c.Flags().String(ConsensusPubKeys, "", "public keys for consensus peers, sep by ','. ")
	_ = c.MarkFlagRequired(EthRpcAddr)

	return c
}

func CreateSyncBscGenesisHdrTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "create_sync_bsc_genesis_hdr_tx [bsc_chain_id] [bsc_epoch_hdr_height]",
		Short: "create transaction to sync bsc header to Poly.",
		RunE:  CreateSyncBscGenesisHdrToPolyTx,
	}

	c.Flags().String(BscRpcAddr, "", "bsc node RPC address")
	c.Flags().String(ConsensusPubKeys, "", "public keys for consensus peers, sep by ','. ")
	_ = c.MarkFlagRequired(BscRpcAddr)

	return c
}

func CreateSyncOkGenesisHdrTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "create_sync_ok_genesis_hdr_tx [ok_chain_id]",
		Short: "create transaction to sync ok header to Poly.",
		RunE:  CreateSyncOkGenesisHdrToPolyTx,
	}

	c.Flags().String(ConsensusPubKeys, "", "public keys for consensus peers, sep by ','. ")

	return c
}

func CreateSyncRawGenesisHdrTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "create_sync_raw_genesis_hdr_tx [chain_id] [raw_hex_file]",
		Short: "create transaction to sync raw genesis header to Poly.",
		RunE:  CreateSyncRawGenesisHdrTxToPolyTx,
	}

	c.Flags().String(ConsensusPubKeys, "", "public keys for consensus peers, sep by ','. ")

	return c
}

func CreateSyncMscGenesisHdrTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "create_sync_msc_genesis_hdr_tx [msc_chain_id] [msc_epoch_hdr_height]",
		Short: "create transaction to sync msc header to Poly.",
		RunE:  CreateSyncMscGenesisHdrToPolyTx,
	}

	c.Flags().String(MscRpcAddr, "", "msc node RPC address")
	c.Flags().String(ConsensusPubKeys, "", "public keys for consensus peers, sep by ','. ")
	_ = c.MarkFlagRequired(MscRpcAddr)

	return c
}

func CreateSyncSwticheoGenesisHdrTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "create_sync_switcheo_genesis_hdr_tx [swth_chain_id] [swth_hdr_height]",
		Short: "create transaction to sync Switcheo header to Poly.",
		RunE:  CreateSyncSwthGenesisHdrToPolyTx,
		Args:  cobra.ExactArgs(2),
	}

	c.Flags().String(SwitcheoRpcAddr, "http://localhost:20336", "Switcheo node RPC address")
	c.Flags().String(ConsensusPubKeys, "", "public keys for consensus peers, sep by ','. ")
	_ = c.MarkFlagRequired(SwitcheoRpcAddr)

	return c
}

func SignPolyMultiSigTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "sign_poly_multisig_tx [raw_tx]",
		Short: "sign multisig transaction of Poly",
		RunE:  SignPolyMultiSigTx,
		Args:  cobra.ExactArgs(1),
	}
	return c
}

func CreateSyncNeoGenesisHdrTxCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "create_sync_neo_genesis_hdr_tx [neo_chain_id] [height]",
		Short: "create transaction to sync NEO header to Poly.",
		RunE:  CreateSyncNeoGenesisHdrTx,
	}

	c.Flags().String(NeoRpcAddr, "", "NEO node RPC address")
	c.Flags().String(ConsensusPubKeys, "", "public keys for consensus peers, sep by ','. ")
	_ = c.MarkFlagRequired(NeoRpcAddr)

	return c
}

func SwitcheoCmd() *cobra.Command {
	sc := &cobra.Command{
		Use:   "switcheo",
		Short: "swticheo subcommands",
		Long: "This command handles all functions about Switcheo, like syncing " +
			"genesis headers of Poly to Switcheo, etc. ",
	}

	sc.PersistentFlags().String(SwitcheoRpcAddr, "", "switcheo rpc address")
	sc.PersistentFlags().String(SwitcheoWallet, "", "switcheo wallet path")
	sc.PersistentFlags().String(SwitcheoWalletPwd, "", "switcheo wallet password")

	sc.AddCommand(
		SyncPolyGenesisHdrToSwitcheoCmd())

	return sc
}

func SyncPolyGenesisHdrToSwitcheoCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "sync_poly_genesis_hdr_to_switcheo [height] [swth_gas] [swth_price]",
		Short: "sync genesis header of poly to switcheo",
		RunE:  SyncPolyHdrToSwitcheo,
		Args:  cobra.ExactArgs(3),
	}

	c.Flags().String(PolyRpcAddr, "", "poly node rpc address")

	return c
}

//func EthereumCmd() *cobra.Command {
//	ec := &cobra.Command{
//		Use:   "switcheo",
//		Short: "swticheo subcommands",
//		Long:  "This command handles all functions about Switcheo, like syncing " +
//			"genesis headers of Poly to Switcheo, etc. ",
//	}
//
//	ec.PersistentFlags().String(EthRpcAddr)
//}

func PolyNSMCmd() *cobra.Command {
	rm := &cobra.Command{
		Use:   "neo3_state_manager",
		Short: "Neo3 state manager subcommands",
		Long:  "neo3 state manager controls neo3 state validators. Like registering or removing a state validator.",
	}
	rm.AddCommand(
		RegisterStateValidatorCmd(),
		ApproveRegisterStateValidatorCmd(),
		RemoveStateValidatorCmd(),
		ApproveRemoveStateValidatorCmd())
	return rm
}

func RegisterStateValidatorCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "register_state_validator [neo3_state_validators_sep_by_comma]",
		Short: "register new neo3 state validators",
		Long:  "You can register new neo3 state validators and toolbox will return the registration ID which is needed when approving.",
		RunE:  RegisterStateValidator,
		Args:  cobra.MinimumNArgs(1),
	}
	return c
}

func ApproveRegisterStateValidatorCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "approve_register_StateValidator [id]",
		Short: "approve registration of proposal ID",
		Long:  "Consensus accounts approve the registration with the ID returned when registering.",
		RunE:  ApproveRegisterStateValidator,
		Args:  cobra.ExactArgs(1),
	}
	return c
}

func RemoveStateValidatorCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "remove_state_validator [neo3_state_validators_sep_by_comma]",
		Short: "remove neo3 state validators",
		Long:  "You can remove neo3 state validators and toolbox will return the removal ID which is needed when removing",
		RunE:  RemoveStateValidator,
	}
	return c
}

func ApproveRemoveStateValidatorCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "approve_remove_state_validator [id]",
		Short: "approve removing proposal ID",
		Long:  "Consensus accounts approve the relayer removing with the ID returned when removing.",
		RunE:  ApproveRemoveStateValidator,
	}
	return c
}
