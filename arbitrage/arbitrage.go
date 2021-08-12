package arbitrage

import (
	"context"
	"cosmos-arbitrage/common"
	"cosmos-arbitrage/config"
	"cosmos-arbitrage/cosmos"
	"cosmos-arbitrage/graph"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
)

func NewArbitrage() Arbitrage {
	return &arbitrage{}
}

type Arbitrage interface {
	Run(ctx context.Context)
}

type arbitrage struct{}

func (a *arbitrage) Run(ctx context.Context) {
	var (
		ticker = time.NewTicker(time.Second)
	)
	client, err := cosmos.NewCosmosClient(ctx, config.CosmosRPCHost)
	if err != nil {
		panic(fmt.Sprintf("NewCosmosClient fail [%s]", err))
	}
	defer client.Close()
	for {
		resp, err := a.runStep(ctx, &RunStepRequest{
			CosmosClient: client,
		})
		if err != nil {
			panic(fmt.Sprintf("runStep fail [%s]", err))
		}

		ticker.Reset(resp.NextTime)
		<-ticker.C
	}
}

type RunStepRequest struct {
	CosmosClient *cosmos.CosmosClient
}

type RunStepResponse struct {
	NextTime time.Duration
}

func (a *arbitrage) runStep(ctx context.Context, request *RunStepRequest) (*RunStepResponse, error) {
	response := &RunStepResponse{
		NextTime: time.Second, // default next step time
	}
	go a.processArbitrage(ctx, &ProcessArbitrageRequest{
		CosmosClient: request.CosmosClient,
	})
	return response, nil
}

type ProcessArbitrageRequest struct {
	CosmosClient *cosmos.CosmosClient
}

func (a *arbitrage) processArbitrage(ctx context.Context, request *ProcessArbitrageRequest) {
	if poolSyncData == nil {
		common.LogErr("poolSyncData is nil wait a second for next step")
		return
	}
	var (
		err       error
		graphPool = []*graph.Pool{}
	)
	// find circle
	poolSyncData.PoolMap.Range(func(key interface{}, value interface{}) bool {
		pool, ok := value.(*PoolData)
		if !ok {
			err = common.Errorf(nil, "value to PoolData fail [%+v: %+v]", key, value)
			return false
		}
		tmp := &graph.Pool{
			ID:     pool.ID,
			Denoms: []string{},
		}
		for denom := range pool.DenomMap {
			tmp.Denoms = append(tmp.Denoms, denom)
		}
		graphPool = append(graphPool, tmp)
		return true
	})
	if err != nil {
		panic(fmt.Sprintf("range PoolMap fail [%s]", err))
	}
	foundPath := graph.FindCircle(graphPool, "uatom", 5)
	// path
	for _, path := range foundPath {
		var (
			rate      = types.NewDec(1)
			tempDenom = "uatom"
		)
		for _, pool := range path.Path {
			poolDataI, ok := poolSyncData.PoolMap.Load(pool.ID)
			if !ok {
				panic(fmt.Sprintf("can not find [%d] in PoolMap fail [%+v]", pool.ID, poolSyncData.PoolMap))
			}
			poolData, ok := poolDataI.(*PoolData)
			if !ok {
				panic(fmt.Sprintf("can not convert [%+v] to PoolData", poolDataI))
			}
			if pool.Denoms[0] == tempDenom {
				rate = rate.Mul(getRate(pool.Denoms[0], pool.Denoms[1], poolData.DenomMap))
				tempDenom = pool.Denoms[1]
			} else {
				rate = rate.Mul(getRate(pool.Denoms[1], pool.Denoms[0], poolData.DenomMap))
				tempDenom = pool.Denoms[0]
			}
		}
		if rate.GT(types.NewDec(1)) {
			common.Log(fmt.Sprintf("!!!!!! found arbitrage circle [%+v] rate [%s]", path.Path, rate))
		}
	}
}

func getRate(from, to string, denomMap map[string]types.Int) types.Dec {
	return types.NewDecFromInt(denomMap[to]).Quo(types.NewDecFromInt(denomMap[from]))
}
