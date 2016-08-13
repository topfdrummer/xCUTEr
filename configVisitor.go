// Copyright (c) 2016 Niklas Wolber
// This file is licensed under the MIT license.
// See the LICENSE file for more information.

package main

import (
	"errors"
	"fmt"
	"log"
	"time"
)

const (
	sequentialFlow = "sequential"
	parallelFlow   = "parallel"

	outputKey     = "output"
	loggerKey     = "logger"
	sshClientKey  = "sshClient"
	templatingKey = "templating"
	stdoutKey     = "stdout"
	stderrKey     = "stderr"
)

type configVisitor interface {
	Sequential() group
	Parallel() group
	Job(name string) group
	JobLogger(jobName string) interface{}
	HostLogger(jobName string, h *host) interface{}
	Timeout(timeout time.Duration) interface{}
	SCP(scp *scp) interface{}
	Host(c *config, h *host) group
	ErrorSafeguard(child interface{}) interface{}
	Templating(c *config, h *host) interface{}
	SSHClient(host, user string) interface{}
	Forwarding(f *forwarding) interface{}
	Commands(cmd *command) group
	Command(cmd *command) interface{}
	Stdout(file string) interface{}
	Stderr(file string) interface{}
}

type group interface {
	Append(children ...interface{})
	Wrap() interface{}
}

func visitConfig(builder configVisitor, c *config) (interface{}, error) {
	children := builder.Job(c.Name)

	if c.Host == nil && c.HostsFile == nil {
		return nil, errors.New("either 'host' or 'hostsFile' must be present")
	}

	if c.Host != nil && c.HostsFile == nil {
		return nil, errors.New("either 'host' or 'hostsFile' may be present")
	}

	children.Append(builder.JobLogger(c.Name))

	if c.Timeout != "" {
		timeout, err := time.ParseDuration(c.Timeout)
		if err != nil {
			return nil, err
		}
		children.Append(builder.Timeout(timeout))
	}

	if c.SCP != nil {
		children.Append(builder.SCP(c.SCP))
	}

	if c.Host != nil {
		host, err := visitHost(builder, c, c.Host)
		if err != nil {
			return nil, err
		}
		children.Append(host)
	}

	if c.HostsFile != nil {
		hosts, err := loadHostsFile(c.HostsFile)
		if err != nil {
			return nil, err
		}

		log.Printf("filtered hosts: %#v", hosts)

		hostFluncs := builder.Parallel()
		for _, host := range *hosts {
			host, err := visitHost(builder, c, host)
			if err != nil {
				return nil, err
			}
			hostFluncs.Append(host)
		}
		children.Append(hostFluncs.Wrap())
	}

	return children.Wrap(), nil
}

func visitHost(builder configVisitor, c *config, host *host) (interface{}, error) {
	if c.Command == nil {
		return nil, errors.New("config does not contain any commands")
	}

	children := builder.Host(c, host)
	children.Append(builder.HostLogger(c.Name, host))
	children.Append(builder.Templating(c, host))
	children.Append(builder.SSHClient(fmt.Sprintf("%s:%d", host.Addr, host.Port), host.User))

	if f := c.Forwarding; f != nil {
		children.Append(builder.Forwarding(f))
	}

	cmd, err := visitCommand(builder, c.Command)
	if err != nil {
		return nil, err
	}
	children.Append(cmd)

	return builder.ErrorSafeguard(children.Wrap()), nil
}

func visitCommand(builder configVisitor, cmd *command) (interface{}, error) {
	const (
		sequential = "sequential"
		parallel   = "parallel"
	)

	if cmd.Command != "" && cmd.Commands != nil && len(cmd.Commands) > 0 {
		err := fmt.Errorf("either command or commands can be present in %s", cmd)
		return nil, err
	}

	var stdout, stderr interface{}
	children := builder.Commands(cmd)

	if cmd.Stdout != "" || cmd.Stderr != "" {
		if cmd.Stdout != "" {
			stdout = builder.Stdout(cmd.Stdout)
		}

		if cmd.Stderr == cmd.Stdout {
			stderr = stdout
		} else if cmd.Stderr != "" {
			stderr = builder.Stderr(cmd.Stderr)
		}
	}
	children.Append(stdout, stderr)

	var cmds interface{}

	if cmd.Command != "" {
		cmds = builder.Command(cmd)
	} else if cmd.Commands != nil && len(cmd.Commands) > 0 {
		var childCommands group

		if cmd.Flow == sequentialFlow {
			childCommands = builder.Sequential()
		} else if cmd.Flow == parallelFlow {
			childCommands = builder.Parallel()
		} else {
			err := fmt.Errorf("unknown flow %q", cmd.Flow)
			log.Println(err)
			return nil, err
		}

		for _, cmd := range cmd.Commands {
			log.Printf("%#v", cmd)
			exec, err := visitCommand(builder, cmd)
			if err != nil {
				return nil, err
			}

			childCommands.Append(exec)
		}

		cmds = childCommands.Wrap()
	} else {
		err := fmt.Errorf("either 'command' or 'commands' has to be specified")
		log.Println(err)

		return nil, err
	}

	children.Append(cmds)

	return children.Wrap(), nil
}