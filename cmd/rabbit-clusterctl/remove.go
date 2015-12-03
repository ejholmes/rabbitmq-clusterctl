package main

import "github.com/codegangsta/cli"

var cmdRemove = cli.Command{
	Name:   "remove",
	Usage:  "Removes this node from the cluster.",
	Action: runRemove,
}

func runRemove(c *cli.Context) {
	ctl := newController(c)
	must(ctl.Remove())
}
