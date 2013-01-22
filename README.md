# WebotsManager

A tool to manage multiple webost installation

## Installation

Install it using gem and specific_install (not already published to rubygem )
    
    $ gem install specific_install
    $ gem install specific_install -l https://github.com/biorob/webots_manager.git


## Usage

    $ webots_manager help

Lists all available task, like :

* list : list installed and available versions
* install VERSION : install a specific version
* use VERSION : use a specific version
* add_template FILENAME WEBOTS_LOCAL_PATH : put the specifioed file in all installed and futurly installed version. Checks options !
* remove_template WEBOTS_LOCAL_PATH : remove the previously template associated to the WEBOTS_LOCAL_PATH on all currently installed version

Remember that each of this task may have specific options. Please use help task to know more about them !

Also remember that this tool is for system wide installation. You will need administrative right (use sudo).

If you tired to use sudo, just remind that all command only modify/write data to the /usr/local/webots_manager directory. You could create a group and change ownership to this group to that directory.

## Bug and Issues 

If you found anything not working, not clear or any unexpected behavior, please report [issues](https://github.com/biorob/webots_manager/issues) on github.

## Contributing

1. Fork it from https://github.com/biorob/webots_manager
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request at github
