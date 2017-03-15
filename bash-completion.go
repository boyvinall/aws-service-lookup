package main

import (
	"fmt"

	"github.com/urfave/cli"
)

var CmdBashCompletion = cli.Command{
	Name:     "bash-completion",
	Usage:    "",
	Action:   bashCompletion,
	Flags:    []cli.Flag{},
	Category: "Misc",
}

func bashCompletion(c *cli.Context) error {
	fmt.Print(`#! /bin/bash

_cli_bash_autocomplete() {
     local cur opts base
     COMPREPLY=()
     cur="${COMP_WORDS[COMP_CWORD]}"
     opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
     COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
     return 0
}

complete -F _cli_bash_autocomplete aws-service-lookup
`)
	return nil
}
