require 'webots_manager/config'

require 'timeout'

require 'open-uri'

module WebotsManager

  class InstanceManager

    def initialize
      restore_state
      get_available_from_archive
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

    def restore_state
      @available = {}
      @installed = []
      @last_update = Time.utc(2000,"jan",1,20,15,1)
      @in_use    = nil
    end

    def save_state
      #not implemented
    end

    def get_available_from_archive
      file = open configatron.cyberbotics_archive , :read_timeout => 10

      if file.last_modified and file.last_modified < @last_update
        return
      end
      escaped_suffix = configatron.suffix.gsub(/\./,'\.')
      r = /[Ww]ebots-([0-9]+(\.[0-9]+)*)-#{configatron.arch}#{escaped_suffix}/
      file.each_line do |l|
        r.match(l) do |m|
          @available[m[1]] = "#{configatron.cyberbotics_archive}#{m}"
        end
      end
    end
  end
end
