package main

import "github.com/codegangsta/cli"

var cmdJoin = cli.Command{
	Name:   "join",
	Usage:  "Joins this node to the cluster",
	Action: runJoin,
}

func runJoin(c *cli.Context) {
	ctl := newController(c)
	must(ctl.Join())
}
