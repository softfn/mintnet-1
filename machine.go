package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	. "github.com/tendermint/go-common"

	"github.com/codegangsta/cli"
)

//--------------------------------------------------------------------------------

func cmdSsh(c *cli.Context) {
	args := c.Args()
	machines := ParseMachines(c.String("machines"))
	cmdBase(args, machines, sshCmd)
}

func cmdScp(c *cli.Context) {
	args := c.Args()
	machines := ParseMachines(c.String("machines"))
	fmt.Println(args, machines)
	cmdBase(args, machines, scpCmd)
}

func cmdBase(args, machines []string, cmd func(string, []string) error) {
	var wg sync.WaitGroup
	for _, mach := range machines {
		wg.Add(1)
		go func(mach string) {
			maybeSleep(len(machines), 2000)
			defer wg.Done()
			cmd(mach, args)
		}(mach)
	}
	wg.Wait()
}

func sshCmd(mach string, args []string) error {
	args = []string{"ssh", mach, strings.Join(args, " ")}
	if !runProcess("ssh-cmd-"+mach, "docker-machine", args, true) {
		return errors.New("Failed to exec ssh command on machine " + mach)
	}
	return nil
}

func scpCmd(mach string, args []string) error {
	if len(args) != 2 {
		return errors.New("scp expects exactly two args")
	}
	args = []string{"scp", args[0], mach + ":" + args[1]}
	if !runProcess("ssh-cmd-"+mach, "docker-machine", args, true) {
		return errors.New("Failed to exec ssh command on machine " + mach)
	}
	return nil
}

//--------------------------------------------------------------------------------

func cmdCreate(c *cli.Context) {
	args := c.Args()
	machines := ParseMachines(c.String("machines"))

	errs := createMachines(machines, args)
	if len(errs) > 0 {
		Exit(Fmt("There were %v errors", len(errs)))
	} else {
		fmt.Println(Fmt("Successfully created %v machines", len(machines)))
	}
}

func createMachines(machines []string, args []string) (errs []error) {
	var wg sync.WaitGroup
	for _, mach := range machines {
		wg.Add(1)
		go func(mach string) {
			maybeSleep(len(machines), 2000)
			defer wg.Done()
			err := createMachine(args, mach)
			if err != nil {
				errs = append(errs, err) // TODO: make thread safe
			}
		}(mach)
	}
	wg.Wait()
	return errs
}

func createMachine(args []string, mach string) error {
	args = append([]string{"create"}, args...)
	args = append(args, mach)
	if !runProcess("create-"+mach, "docker-machine", args, true) {
		return errors.New("Failed to create machine " + mach)
	}
	return nil
}

//--------------------------------------------------------------------------------

func cmdDestroy(c *cli.Context) {
	machines := ParseMachines(c.String("machines"))

	// Destroy each machine.
	var wg sync.WaitGroup
	for _, mach := range machines {
		wg.Add(1)
		go func(mach string) {
			defer wg.Done()
			err := removeMachine(mach)
			if err != nil {
				fmt.Println(Red(err.Error()))
			}
		}(mach)
	}
	wg.Wait()

	fmt.Println("Success!")
}

//--------------------------------------------------------------------------------

func cmdProvision(c *cli.Context) {
	args := c.Args()
	machines := ParseMachines(c.String("machines"))

	errs := provisionMachines(machines, args)
	if len(errs) > 0 {
		Exit(Fmt("There were %v errors", len(errs)))
	} else {
		fmt.Println(Fmt("Successfully created %v machines", len(machines)))
	}
}

func provisionMachines(machines []string, args []string) (errs []error) {
	var wg sync.WaitGroup
	for _, mach := range machines {
		wg.Add(1)
		go func(mach string) {
			maybeSleep(len(machines), 2000)
			defer wg.Done()
			err := provisionMachine(args, mach)
			if err != nil {
				errs = append(errs, err)
			}
		}(mach)
	}
	wg.Wait()
	return errs
}

func provisionMachine(args []string, mach string) error {
	args = append([]string{"provision"}, args...)
	args = append(args, mach)
	if !runProcess("provision-"+mach, "docker-machine", args, true) {
		return errors.New("Failed to provision machine " + mach)
	}
	return nil
}

//--------------------------------------------------------------------------------

// Stop a machine
// mach: name of machine
func stopMachine(mach string) error {
	args := []string{"stop", mach}
	if !runProcess("stop-"+mach, "docker-machine", args, true) {
		return errors.New("Failed to stop machine " + mach)
	}
	return nil
}

// Remove a machine
// mach: name of machine
func removeMachine(mach string) error {
	args := []string{"rm", "-f", mach}
	if !runProcess("remove-"+mach, "docker-machine", args, true) {
		return errors.New("Failed to remove machine " + mach)
	}
	return nil
}

// List machine names that match prefix
func listMachines(prefix string) ([]string, error) {
	args := []string{"ls", "--quiet"}
	output, ok := runProcessGetResult("list-machines", "docker-machine", args, true)
	if !ok {
		return nil, errors.New("Failed to list machines")
	}
	output = strings.TrimSpace(output)
	if len(output) == 0 {
		return nil, nil
	}
	machines := strings.Split(output, "\n")
	matched := []string{}
	for _, mach := range machines {
		if strings.HasPrefix(mach, prefix+"-") {
			matched = append(matched, mach)
		}
	}
	return matched, nil
}

// Get ip of a machine
// mach: name of machine
func getMachineIP(mach string) (string, error) {
	args := []string{"ip", mach}
	output, ok := runProcessGetResult("get-ip-"+mach, "docker-machine", args, true)
	if !ok {
		return "", errors.New("Failed to get ip of machine" + mach)
	}
	return strings.TrimSpace(output), nil
}
