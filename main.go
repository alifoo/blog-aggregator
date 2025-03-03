package main

import (
	"fmt"

	"github.com/alifoo/blog-aggregator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cfg)

	cfg.SetUser("alifoo")

	cfg, err = config.Read()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg)
}
