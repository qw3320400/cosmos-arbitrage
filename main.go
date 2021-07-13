package main

import (
	"context"
	"cosmos-arbitrage/arbitrage"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// signal chan
	c := make(chan os.Signal)
	// listen signal
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

	ctx := context.Background()

	syncer := arbitrage.NewPoolSyncer()
	go syncer.Sync(ctx)

	arb := arbitrage.NewArbitrage()
	go arb.Run(ctx)

	select {
	case sig := <-c:
		fmt.Println("recieve signal", sig, "process close")
		return
	}
}
