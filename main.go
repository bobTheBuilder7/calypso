package main

import (
	"cmp"
	"context"
	"encoding/json/v2"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [filepath]\n", os.Args[0])
		os.Exit(1)
	}

	filepath := ""
	if len(os.Args) > 1 {
		filepath = os.Args[1]
	}
	filepath = cmp.Or(filepath, "./config.json")

	configFile, err := os.Open(filepath)
	if err != nil {
		panic(err.Error())
	}

	var cfg Config

	err = json.UnmarshalRead(configFile, &cfg)
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()

	run(ctx, cfg)
}
