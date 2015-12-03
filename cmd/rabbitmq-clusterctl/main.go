package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/remind101/rabbitmq-clusterctl"
)

var commands = []cli.Command{
	cmdMaster,
	cmdPromote,
	cmdJoin,
	cmdRemove,
}

func main() {
	app := cli.NewApp()
	app.Name = "rabbitmq-clusterctl"
	app.Usage = "Perform rabbitmq node operations"
	app.Commands = commands
	app.Run(os.Args)
}

func newController(c *cli.Context) *clusterctl.Controller {
	hostname, _ := os.Hostname()

	return &clusterctl.Controller{
		Node:                 fmt.Sprintf("rabbit@%s", hostname),
		MasterController:     clusterctl.NewELBMasterController(os.Getenv("ELB_NAME")),
		MembershipController: clusterctl.DefaultMembershipController,
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
