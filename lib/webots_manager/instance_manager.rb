require 'webots_manager/config'

require 'timeout'

require 'open-uri'

module WebotsManager

  class InstanceManager

    def initialize
      restore_state
    end

    attr_reader :available

    attr_reader :installed

    attr_reader :in_use

    def installed? version
      @installed.include? version
    end

    def in_use? version
      version === in_use
    end

    def install
      raise "Unimplemented yet"
    end

    def remove
      raise "Unimplemented yet"
    end

    def use
      raise "Unimplemented yet"
    end

    private

    def create_if_needed_wdir
      @wdir = configatron.install_prefix
      Dir.mkdir(@wdir,0755) unless File.directory? @wdir
    end

    def restore_state
      create_if_needed_wdir
      get_available_from_archive
      get_installed_from_wdir
      get_in_use_from_wdir
    end

    def get_available_from_archive
      @available = {}
      file = open configatron.cyberbotics_archive , :read_timeout => 10

      escaped_suffix = configatron.suffix.gsub(/\./,'\.')
      r = /[Ww]ebots-([0-9]+(\.[0-9]+)*)-#{configatron.arch}#{escaped_suffix}/
      file.each_line do |l|
        r.match(l) do |m|
          @available[m[1]] = "#{configatron.cyberbotics_archive}#{m}"
        end
      end

    end

    def get_installed_from_wdir
      @installed = []
      Dir.new(@wdir).each do |f|
        /[0-9]+(\.[0-9]+)*/.match(f) do |m|
          @installed.insert(m)
          puts "Version #{m} installed"
        end
      end
    end

    def get_in_use_from_wdir
      @in_use = nil
      Dir.chdir(@wdir) do
        if File.symlink?('in_use')
          @in_use = File.readlink('in_use')
        end
      end
    end


  end
end
