require 'webots_manager/version'
require 'thor'

require 'webots_manager/config'

module WebotsManager

  class CliInterface < Thor

    desc "version" , "print the current version number"
    def version
      puts "webots_manager #{VERSION}"
    end

  end
end
