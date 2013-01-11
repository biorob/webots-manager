require 'webots_manager/config'

module WebotsManager

  class InstanceManager

    def initialize
    end

    def available
      []
    end

    def installed
      []
    end

    def installed? version
      false
    end

    def in_use
      nil
    end

    def in_use? version
      false
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

  end
end
