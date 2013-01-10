require 'webots_manager/version'
require 'thor'

module WebotsManager

  class CliInterface < Thor
    desc "version" , "print the current version number"
    def version
      puts "webots_manager #{VERSION}"
    end
  end
end
