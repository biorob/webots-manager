package main

import (
	"fmt"
	"log"
)

type ListCommand struct {
	All bool `short:"a" long:"available" description:"also prints all available version for installation"`
}

type InitCommand struct{}

type InstallCommand struct {
	Use bool `short:"u" long:"use" description:"force use of this new version after installation"`
}

type Interactor struct {
	archive WebotsArchive
	manager WebotsInstanceManager
}

func NewInteractor() (*Interactor, error) {
	res := &Interactor{}
	var err error
	res.archive, err = NewWebotsHttpArchive("http://www.cyberbotics.com/archive/")
	if err != nil {
		return nil, err
	}
	res.manager, err = NewSymlinkManager(res.archive)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (x *ListCommand) Execute(args []string) error {
	xx, err := NewInteractor()
	if err != nil {
		return err
	}
	installed := xx.manager.Installed()
	if len(installed) == 0 {
		fmt.Printf("No webots version installed.\n")
	} else {
		for _, v := range installed {
			if xx.manager.IsUsed(v) == true {
				fmt.Printf(" -* %s\n", v)
			} else {
				fmt.Printf(" -  %s\n", v)
			}
		}
	}
	if x.All {
		fmt.Println("List of all available versions:")
		for _, v := range xx.archive.AvailableVersions() {
			fmt.Printf(" - %s\n", v)
		}
	} else {
		vers := xx.archive.AvailableVersions()
		if len(vers) == 0 {
			return fmt.Errorf("No version are available")
		}
		fmt.Printf("Last available version is %s\n",
			vers[len(vers)-1])
	}

	return nil
}

func (x *InitCommand) Execute(args []string) error {
	return SymlinkManagerSystemInit()
}

func (x *InstallCommand) Execute(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Missing version to install")
	}

	v, err := ParseWebotsVersion(args[0])
	if err != nil {
		return err
	}

	xx, err := NewInteractor()
	if err != nil {
		return err
	}

	err = xx.manager.Install(v)
	if err != nil {
		return err
	}

	notUsed := true
	for _, vv := range xx.manager.Installed() {
		if xx.manager.IsUsed(vv) {
			notUsed = false
			break
		}
	}

	if notUsed || x.Use {
		err = xx.manager.Use(v)
		if err != nil {
			return err
		}
		log.Printf("Using now version %s", v)
	}
	return nil
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
