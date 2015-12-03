package main

import "github.com/codegangsta/cli"

var cmdPromote = cli.Command{
	Name:   "promote",
	Usage:  "Promotes this node to be the new master.",
	Action: runPromote,
}

func runPromote(c *cli.Context) {
	ctl := newController(c)
	must(ctl.Promote())
}
