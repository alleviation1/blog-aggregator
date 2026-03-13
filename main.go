package main

import (
	"fmt"

	"github.com/alleviation1/blog_aggregator/internal/config"
)

const configFileName = "/gatorconfig.json"

func main() {
	cfg, err := config.Read(configFileName)
	if err != nil {
		fmt.Errorf("Error creating config: %w", err)
		return
	}
	fmt.Println(cfg.Url)
	err = cfg.SetUser("Alec")
	if err != nil {
		fmt.Errorf("Error creating config: %w", err)
		return
	}

	fmt.Println(config.Read(configFileName))
}