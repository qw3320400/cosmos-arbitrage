package cosmos

import (
	"context"
	"cosmos-arbitrage/common"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	banktype "github.com/cosmos/cosmos-sdk/x/bank/types"
	liqtype "github.com/gravity-devs/liquidity/x/liquidity/types"
	"google.golang.org/grpc"
)

func NewCosmosClient(ctx context.Context, host string) (*CosmosClient, error) {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return nil, common.Errorf(err, "grpc.Dial fail host [%s]", host)
	}
	cli := &CosmosClient{
		conn:             conn,
		TendermintClient: tmservice.NewServiceClient(conn),
		LiquidityClient:  liqtype.NewQueryClient(conn),
		BankClient:       banktype.NewQueryClient(conn),
	}
	return cli, nil
}

type CosmosClient struct {
	conn             *grpc.ClientConn
	TendermintClient tmservice.ServiceClient
	LiquidityClient  liqtype.QueryClient
	BankClient       banktype.QueryClient
}

func (c *CosmosClient) Close() {
	c.conn.Close()
}
