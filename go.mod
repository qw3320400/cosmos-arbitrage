module cosmos-arbitrage

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.42.6
	github.com/gravity-devs/liquidity v1.2.9
	github.com/smartystreets/assertions v1.1.1 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	google.golang.org/grpc v1.38.0
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
