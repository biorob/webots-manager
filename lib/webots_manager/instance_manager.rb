require 'webots_manager/config'

require 'timeout'

require 'open-uri'

require 'ruby-progressbar'


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

    def install version
      if @installed.include? version
        raise "#{version} is already installed"
      end

      unless @available.include? version
        raise "#{version} is not available"
      end


      tmpfile = Tempfile.new(['webots_manager_download-',
                               configatron.suffix],
                         Dir.tmpdir,'rwb')

      puts "Downloading version #{version} in #{tmpfile.path}"
      pbar = nil

      open(@available[version],
      :content_length_proc => lambda { |t|
             if t && 0 < t
               pbar = ProgressBar.create :total => t
             end
           },
      :progress_proc => lambda { |s|
             pbar.progress = s if pbar }) do |f|
        tmpfile.write(f.read)
      end


      install_path = File.join(@wdir, version + "_in_installation")
      final_path = File.join(@wdir,version)
      puts "Extracting #{tmpfile.path} to #{install_path}"
      Dir.chdir(@wdir) do
        d = Dir.mkdir(install_path) unless File.directory? install_path
        Dir.chdir(install_path) do
          system("tar -x -f #{tmpfile.path} -C #{install_path}") 
        end
        webots_install = File.join(install_path,'webots')
        puts "Moving #{webots_install} to #{final_path}"
        File.rename(webots_install,final_path)
        Dir.rmdir(install_path) # should be empty
      end
      @installed.push version
    end

    def remove
      raise "Unimplemented yet"
    end

    def use version
      unless @installed.include? version
        raise "Version #{version} is not installed"
      end

      Dir.chdir(@wdir) do

        if @in_use
          File.delete('in_use')
          @in_use = nil
        end

        File.symlink(version,'in_use')
        @in_use = version

      end

      #test if global symlink is ok
      Dir.chdir(File.dirname(@wdir)) do
        if not File.symlink?('webots')
          File.symlink('webots_manager/in_use','webots')
        elsif File.readlink('webots') != 'webots_manager/in_use'
          raise "Symlink to webots does not point to webots_manager/in_use"
        end
      end

      #Test if Webots Home is set or not
      if ENV['WEBOTS_HOME'].nil?
        puts "WEBOTS_HOME environment variable is not set. please add :
export WEBOTS_HOME=/usr/local/webots to your .bashrc or .profile"
      elsif ENV['WEBOTS_HOME'] != '/usr/local/webots'
        puts "WARNING : you should change your WEBOTS_HOME variable to point to
/usr/local/webots, it is currently #{ENV['WEBOTS_HOME']}"
      end

    end


    def self.init
      unless Process::uid == 0 && Process::euid == 0
        raise "I must be run by root for this operation"
      end

      gname = configatron.group_name

      begin 
        rgid = Process::GID.from_name(gname)
      rescue
        system "addgroup  #{gname}"
        rgid = Process::GID.from_name(gname)
      end

      wdir = configatron.install_prefix



      #Adds a global symlink
      Dir.chdir(File.dirname(wdir)) do

        if File.exists?('webots')|| File.symlink?('webots')
          File.rename 'webots', 'webots.webots_manager.bak'
        end        
        File.symlink('webots_manager/in_use','webots')
      end


      Dir.mkdir(wdir,755)
      FileUtils.chown 'root', gname, wdir
      FileUtils.chmod "ug=rwxs,o=rx", wdir
      puts "Working directory have been created in #{wdir}, and group '#{gname}' have been created. To use webots_manager, please add yourself to the group '#{gname}'"

      ## now evrything is ok !
    end

    def is_init?
      @wdir = configatron.install_prefix
      return File.directory? @wdir
    end

    private

    
    
    def restore_state
      unless is_init?
        raise "Installation is not initialized, please use init command first"
      end
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
          @installed.push(f)
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
