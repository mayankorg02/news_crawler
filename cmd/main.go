package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"news-crawler/internal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := internal.InitRunner().Run(ctx, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
