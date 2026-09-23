package main

import (
	"context"
	"craft/internal/cli"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := cli.Main(ctx)
	stop()
	os.Exit(code)
}
