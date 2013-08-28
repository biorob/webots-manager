require 'webots_manager/version'
require 'thor'

require 'webots_manager/config'
require 'webots_manager/instance_manager.rb'
require 'webots_manager/template_files_manager.rb'


module WebotsManager

  class CliInterface < Thor

    desc "init", "Init webots_manager. Shoudl be runned as root"
    def init
      InstanceManager.init
    end


    desc "version" , "print the current version number"
    def version
      puts "webots_manager #{VERSION}"
    end

    desc "list" , "print the list of available webots version"
    method_option :all , :type => :boolean, :desc => 'display all available version' , :aliases => '-a'
    def list
      @i = InstanceManager.new
      if @i.installed.empty?
        puts "No version installed yet."
      else
        puts "Installed version :"
        @i.installed.each do |wi|
          puts " -" + (@i.in_use?(wi) ? "* " : "  " ) + wi
        end
      end

      if options[:all]
        puts "Available versions :"
        @i.available.each do |wi,url|
          puts " - " + wi
        end
      elsif not @i.available.empty?
        puts "Latest available version : " + @i.available.keys.last
      else
        puts "No available versions."
      end
    end

    desc "install VERSION", "Install a new webots version, and use it if none is in use"
    def install version
      @i = InstanceManager.new
      unless @i.available.include?(version)
        raise "#{version} is not available"
      end

      if @i.installed.include?(version)
        raise "#{version} is already installed, maybe you want to use it"
      end

      @i.install version
      if @i.in_use.nil?
        @i.use version
      end

      @f = TemplateFilesManager.new
      @f.update_links @i.installed

    end

    desc "use VERSION" , "Ask the system to use webots version VERSION"
    def use version
      @i = InstanceManager.new
      @i.use version
    end

    desc "add_template FILENAME LOCAL_PATH", "Add a template file FILENAME in all $WEBOTS_HOME/LOCAL_PATH"
    method_option :only, :aliases => "-o", :type => :array, :desc => "a list of version to whitelist"
    method_option :except, :aliases => "-e", :type => :array, :desc => "a list of version to blacklist"
    def add_template filename, local_path
      @i = InstanceManager.new
      @f = TemplateFilesManager.new
      @f.add_file filename,local_path,options
      @f.update_links @i.installed
    end

    desc 'remove_template LOCAL_PATH', "Removes all templates that points to $WEBOTS/LOCAL_PATH"
    def remove_template local_path
      @i = InstanceManager.new
      @f = TemplateFilesManager.new
      @f.remove_file local_path, @i.installed
    end
  end
end
