package main

import (
	"github.com/skan0779/skan_coin_2022/cli"
	"github.com/skan0779/skan_coin_2022/db"
)

func main() {
	defer db.Close()
	cli.Start()
}
