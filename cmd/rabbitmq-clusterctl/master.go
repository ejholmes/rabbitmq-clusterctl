package main

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var cmdMaster = cli.Command{
	Name:   "master",
	Usage:  "Outputs the hostname of the current master node, as determined by the ELB.",
	Action: runMaster,
}

func runMaster(c *cli.Context) {
	ctl := newController(c)
	node, err := ctl.Master()
	must(err)
	fmt.Println(node)
}
