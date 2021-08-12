package arbitrage

import (
	"context"
	"cosmos-arbitrage/common"
	"cosmos-arbitrage/config"
	"cosmos-arbitrage/cosmos"
	"cosmos-arbitrage/graph"
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
			PoolMap: SyncMap{},
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
		PoolMap: SyncMap{},
	}
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
			ID:        pool.Id,
			DenomMap:  denomMap,
			DenomList: pool.ReserveCoinDenoms,
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
	PoolMap SyncMap
}

type SyncMap struct {
	sync.Map // map[uint64]*PoolData
}

func (m *SyncMap) MarshalJSON() ([]byte, error) {
	tmpMap := make(map[uint64]*PoolData)
	m.Range(func(k, v interface{}) bool {
		tmpMap[k.(uint64)] = v.(*PoolData)
		return true
	})
	return json.Marshal(tmpMap)
}

type PoolData struct {
	ID        uint64
	DenomMap  map[string]types.Int
	DenomList []string
}

func (p *PoolData) GetID() uint64 {
	return p.ID
}

func (p *PoolData) GetFirstDenom() string {
	if len(p.DenomList) != 2 {
		panic(common.Errorf(nil, "error pool data [%+v]", p))
	}
	return p.DenomList[0]
}

func (p *PoolData) GetSecondDenom() string {
	if len(p.DenomList) != 2 {
		panic(common.Errorf(nil, "error pool data [%+v]", p))
	}
	return p.DenomList[1]
}

var (
	_ graph.Pool = &PoolData{}

	poolSyncData *PoolSyncData
)
