require 'webots_manager/version'
require 'thor'

require 'webots_manager/config'
require 'webots_manager/instance_manager.rb'

module WebotsManager

  class CliInterface < Thor

    desc "version" , "print the current version number"
    def version
      puts "webots_manager #{VERSION}"
    end

    desc "list" , "print the list of available webots version"
    method_option :all , :type => :boolean, :desc => 'display all available version' , :aliases => '-a'
    def list
      instances = InstanceManager.new
      if instances.installed.empty?
        puts "No version installed yet."
      else
        puts "Installed version :"
        instances.installed.each do |wi|
          puts " - " + instances.in_use?(wi) ? "*" : " " + wi
        end
      end

      if options[:all]
        puts "Available versions :"
        instances.available.each do |wi,url|
          puts " - " + wi
        end
      elsif not instances.available.empty?
        puts "Latest available version : " + instances.available.keys.last
      else
        puts "No available versions."
      end
    end
  end
end
