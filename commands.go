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

type UseCommand struct{}

type Interactor struct {
	archive   WebotsArchive
	manager   WebotsInstanceManager
	templates TemplateManager
}

func NewInteractor() (*Interactor, error) {
	res := &Interactor{}
	var err error
	res.archive, err = NewWebotsHttpArchive("http://www.cyberbotics.com/archive/")
	if err != nil {
		return nil, err
	}

	manager, err := NewSymlinkManager(res.archive)
	if err != nil {
		return nil, err
	}
	res.manager = manager
	res.templates = manager.templates

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

func (x *UseCommand) Execute(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Missing version to use")
	}

	v, err := ParseWebotsVersion(args[0])
	if err != nil {
		return err
	}

	xx, err := NewInteractor()
	if err != nil {
		return err
	}

	return xx.manager.Use(v)
}

type AddTemplateCommand struct {
	Only   []string `short:"o" long:"only" description:"apply template only for these versions"`
	Except []string `short:"e" long:"except" description:"do not apply template on these versions"`
}

func (x *AddTemplateCommand) Execute(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Need file to read and where to install")
	}

	var white, black []WebotsVersion
	for _, w := range x.Only {
		v, err := ParseWebotsVersion(w)
		if err != nil {
			return err
		}
		white = append(white, v)
	}

	for _, w := range x.Except {
		v, err := ParseWebotsVersion(w)
		if err != nil {
			return err
		}
		black = append(black, v)
	}

	xx, err := NewInteractor()
	if err != nil {
		return err
	}

	err = xx.templates.RegisterTemplate(args[0], args[1])
	if err != nil {
		return err
	}

	err = xx.templates.WhiteList(args[1], white)
	if err != nil {
		return err
	}
	err = xx.templates.BlackList(args[1], black)
	if err != nil {
		return err
	}

	return xx.manager.ApplyAllTemplates()
}

type RemoveTemplateCommand struct{}

func (x *RemoveTemplateCommand) Execute(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Need install path to remove template from")
	}

	xx, err := NewInteractor()
	if err != nil {
		return err
	}

	err = xx.templates.RemoveTemplate(args[0])
	if err != nil {
		return err
	}

	return xx.manager.ApplyAllTemplates()
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

	parser.AddCommand("use",
		"Use a webots version on the system",
		"Use a webots version on the system. If it is not installed, it will first install it",
		&UseCommand{})

	parser.AddCommand("add-template",
		"Adds a template file to all version",
		"Install a file to all version of webots. -o and -e can be used to explicitely whitelist or blacklist a version",
		&AddTemplateCommand{})

	parser.AddCommand("remove-template",
		"Removes a template file from all version",
		"Removes a previously installed template from all version of webots.",
		&RemoveTemplateCommand{})

}
