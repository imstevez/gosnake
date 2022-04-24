package main

import (
	"context"
	"flag"
	"fmt"
	"gosnake"
	"os"
)

var (
	server bool
)

func init() {
	flag.BoolVar(&server, "srv", false, "start as server")
	flag.StringVar(&(gosnake.DefaultServerOptions.Addr), "listen-addr", "127.0.0.1:9001", "server listen address")
	flag.StringVar(&(gosnake.DefaultClientOptions.ServerAddr), "server-addr", "127.0.0.1:9001", "server address")
}

func main() {
	flag.Parse()
	ctx := context.Background()
	if server {
		err := gosnake.RunServer(ctx)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	err := gosnake.RunClient(ctx)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
