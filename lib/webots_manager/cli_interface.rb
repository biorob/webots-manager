require 'webots_manager/version'
require 'thor'

module WebotsManager

  class CliInterface < Thor
    desc "version" , "print the current version number"
    def version
      puts VERSION
    end
  end
end
