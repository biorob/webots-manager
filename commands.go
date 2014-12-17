package main

import "fmt"

type ListCommand struct {
	All bool `short:"a" long:"available" description:"also prints all available version for installation"`
}

type InitCommand struct{}

type InstallCommand struct{}

func (x *ListCommand) Execute(args []string) error {
	return fmt.Errorf("Not yet implemented")
}

func (x *InitCommand) Execute(args []string) error {
	return fmt.Errorf("Not yet implemented")
}

func (x *InstallCommand) Execute(args []string) error {
	return fmt.Errorf("Not yet implemented")
}

func init() {
	parser.AddCommand("list",
		"Prints all the available version of webots",
		"Prints all installed version, and current version in use. Can also prinst all available version for installation",
		&ListCommand{})

	parser.AddCommand("init",
		"Initialiaze the system for webots_manager",
		"Initialiaze the system with all requirement for webots_manager",
		&InitCommand{})

	parser.AddCommand("install",
		"Install a new webots version on the system",
		"Installs a new webots version on the system",
		&InstallCommand{})

}
