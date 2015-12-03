package clusterctl

import (
	"os"
	"os/exec"
)

type rabbitmqctlFunc func(node string, command string, arg ...string) error

// rabbitmqctl is a function that invokes the rabbitmqctl command using the exec
// package.
func rabbitmqctl(node string, command string, arg ...string) error {
	cmd := exec.Command("rabbitmqctl", append([]string{command}, arg...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
