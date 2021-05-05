package main

import (
	"flag"
	"os"

	"github.com/tsunekawa/meroku/cmd"
)

func main() {
	flag.Parse()
	args := flag.Args()

	switch args[0] {
	case "download":
		cmd.DownloadCmd(args[1:])
	case "parse":
		cmd.ParseCmd(args[1:])
	default:
		flag.Usage()
		os.Exit(1)
	}
}
