package main

import (
	"BIP_backend/internal/app/apiserver"
	"flag"
	"fmt"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/apiserver.toml", "path to config file")
}

func main() {
	fmt.Println("main")
	apiserver.Print()
	apiserver.Print2()
}
