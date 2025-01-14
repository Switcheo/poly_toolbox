module github.com/ontio/poly_toolbox

go 1.14

require (
	github.com/cosmos/cosmos-sdk v0.39.1
	github.com/ethereum/go-ethereum v1.10.16
	github.com/joeqian10/neo-gogogo v1.3.0
	github.com/joeqian10/neo3-gogogo v1.1.2
	github.com/ontio/ontology-crypto v1.0.9
	github.com/ontio/ontology-go-sdk v1.11.4
	github.com/polynetwork/cosmos-poly-module v0.0.0-20200810030259-95d586518759
	github.com/polynetwork/poly v1.8.3
	github.com/polynetwork/poly-go-sdk v0.0.0-20220310062143-c991755afe5f
	github.com/spf13/cobra v1.1.1
	github.com/switcheo/tendermint v0.34.14-2
	github.com/tendermint/tendermint v0.33.7
)

replace github.com/polynetwork/poly-go-sdk => github.com/joeqian10/poly-go-sdk v0.0.0-20210517072349-71002ebfdf13

replace github.com/tendermint/tm-db/064 => github.com/tendermint/tm-db v0.6.4
