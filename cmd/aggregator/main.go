package main

import (
	"github.com/klyakssa/aggregation-sub/internal/app"
	"github.com/klyakssa/aggregation-sub/internal/config"
)

func main() {
	config := config.InitConfiguration()

	app.Run(config)
}
