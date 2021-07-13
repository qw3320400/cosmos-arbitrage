package arbitrage

import (
	"context"
	"cosmos-arbitrage/common"
	"cosmos-arbitrage/config"
	"cosmos-arbitrage/cosmos"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types"
	banktype "github.com/cosmos/cosmos-sdk/x/bank/types"
	liqtype "github.com/gravity-devs/liquidity/x/liquidity/types"
)

func NewPoolSyncer() PoolSyncer {
	return &poolSyncer{}
}

type PoolSyncer interface {
	Sync(ctx context.Context)
}

type poolSyncer struct {
}

func (p *poolSyncer) Sync(ctx context.Context) {
	var (
		ticker = time.NewTicker(time.Second)
	)
	client, err := cosmos.NewCosmosClient(ctx, config.CosmosRPCHost)
	if err != nil {
		panic(fmt.Sprintf("NewCosmosClient fail [%s]", err))
	}

	defer client.Close()
	for {
		resp, err := p.runSync(ctx, &RunSyncRequest{
			CosmosClient: client,
		})
		if err != nil {
			panic(fmt.Sprintf("runSync fail [%s]", err))
		}

		ticker.Reset(resp.NextTime)
		<-ticker.C
	}
}

type RunSyncRequest struct {
	CosmosClient *cosmos.CosmosClient
}

type RunSyncResponse struct {
	NextTime time.Duration
}

func (p *poolSyncer) runSync(ctx context.Context, request *RunSyncRequest) (*RunSyncResponse, error) {
	response := &RunSyncResponse{
		NextTime: time.Second, // default next step time
	}
	if poolSyncData == nil {
		poolSyncData = &PoolSyncData{
			Height:  0,
			PoolMap: sync.Map{},
		}
	}
	blockResp, err := request.CosmosClient.TendermintClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	if err != nil {
		common.LogErr(fmt.Sprintf("GetBlockByHeight fail [%s] wait a second for next step", err))
		return response, nil
	}
	if poolSyncData.Height >= blockResp.Block.Header.Height {
		return response, nil
	}
	// start sync
	common.Log(fmt.Sprintf("syncing liquidity pools at height [%d]", blockResp.Block.Header.Height))
	liqResponse, err := request.CosmosClient.LiquidityClient.LiquidityPools(ctx, &liqtype.QueryLiquidityPoolsRequest{})
	if err != nil {
		common.LogErr(fmt.Sprintf("LiquidityPools fail [%s] wait a second for next step", err))
		return response, nil
	}
	newPoolSyncData := &PoolSyncData{
		Height:  blockResp.Block.Header.Height,
		PoolMap: sync.Map{},
	}
	common.Log(fmt.Sprintf("test log %+v", liqResponse))
	for _, pool := range liqResponse.Pools {
		if len(pool.ReserveCoinDenoms) != 2 {
			return nil, common.Errorf(nil, "ReserveCoinDenoms len error [%+v]", pool.ReserveCoinDenoms)
		}
		balResp, err := request.CosmosClient.BankClient.AllBalances(ctx, &banktype.QueryAllBalancesRequest{
			Address: pool.ReserveAccountAddress,
		})
		if err != nil {
			common.LogErr(fmt.Sprintf("AllBalances fail [%s] wait a second for next step", err))
			return response, nil
		}
		denomMap := map[string]types.Int{}
		for _, denom := range pool.ReserveCoinDenoms {
			var found bool
			for _, bal := range balResp.Balances {
				if bal.Denom == denom {
					found = true
					denomMap[bal.Denom] = bal.Amount
					break
				}
			}
			if !found {
				return nil, common.Errorf(nil, "can not find balance [%+v] [%s] from reserve account [%s]", balResp.Balances, denom, pool.ReserveAccountAddress)
			}
		}
		newPoolData := &PoolData{
			ID:       pool.Id,
			DenomMap: denomMap,
		}
		newPoolSyncData.PoolMap.Store(pool.Id, newPoolData)
	}
	poolSyncData = newPoolSyncData
	common.Log(fmt.Sprintf("sync liquidity pools at height finish [%d]", blockResp.Block.Header.Height))
	poolSyncDataByte, _ := json.Marshal(poolSyncData)
	common.Log(string(poolSyncDataByte))
	return response, nil
}

type PoolSyncData struct {
	Height  int64
	PoolMap sync.Map
}

type PoolData struct {
	ID       uint64
	DenomMap map[string]types.Int
}

type DenomData struct {
	Denom string
}

var (
	poolSyncData *PoolSyncData
)
