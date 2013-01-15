require 'webots_manager/config'
require 'digest/md5'

module WebotsManager
  class TemplateFilesManager

    def initialize
      load_state
    end

    def remove_file filename
      d = compute_digest filename
      unless @templates.include? d
        raise "No file with content of '#{filename}' is stored."
      end
      @templates.delete d
    end

    def add_file filename,path,*opts
      digest = compute_digest filename

      unless @templates.include? digest
        store filename , digest
      end

      unless entry[:paths].include? path
        @templates[digest][:paths].push path
      end

      if opts.include? :only
        @templates[digest][:only] = opts[:only]
      end

      if opts.include? :except
        @templates[digest][:except] = opts[:except]
      end
      save_state
    end

    def has? filename
      @templates.include?(compute_digest(filename))
    end

    def whitelist filename, version
      d = compute_digest filename
      unless @templates.include? d
        raise "File '#{filename}' is not stored"
      end
      unless @templates[d][:only].includes? version
        @templates[d][:only].push version
      end
      save_state
    end


    def blacklist filename, version
      d = compute_digest filename
      unless @templates.include? d
        raise "File '#{filename}' is not stored"
      end
      unless @templates[d][:except].includes? version
        @templates[d][:except].push version
      end
      save_state
    end

    def update_links versions
      @templates.each do |d,opt|
        versions.each do |v|
          should_link = (opt[:only].empty? || opt[:only].include?(v)) && not opt[:except].include?(v)
          if should_link
            create_or_update_link d,opt,v
          else
            remove_link_if_needed opt,v
          end
        end
      end
    end


    private

    def dir_for_version v
      File.join(File.dirname(@wdir),v)
    end

    def is_a_managed_symlink p
      File.symlink? p and File.dirname(File.readlink(p)) == @wdir
    end

    def remove_link_if_needed opt,version
      Dir.chir(dir_for_version(version)) do
        opt[:paths].each do |p|
          if is_a_managed_symlink(p)
            File.unlink(p)
          end
        end
      end
    end

    def create_or_update_link d,opt,version
      target_file = File.join(@wdir,d)
      Dir.chdir(dir_for_version(version)) do
        opt[:paths].each do |p|
          if is_a_managed_symlink(p)
            File.unlink(p)
          end
          unless File.exists? p
            File.symlink(target_file,p)
          end
        end
      end
    end

    def store filename , digest
      stored_entry = File.join(@wdir,digest)
      File.open(filename,'r') do |input|
        File.open(stored_entry,'w') do |output|
          output.write(input.read)
        end
      end
      @templates[digest] = {:paths => [], :only => [], :except => [] }
    end

    def compute_digest filename
      return Digest::MD5.hexdigest(File.read(filename))
    end

    def load_state
      @wdir = File.join(configatron.install_prefix,'templates')
      Diro.mkdir(@wdir,0755) unless File.directory? @wdir
      Dir.chdir(@wdir) do
        @dbfile = 'template_db.yml'
        unless File.exists? @dbfile
          File.open(@dbfile,'w') do |f|
            f.puts ""
          end
        end
        File.open(File.join('templates','template_db.yml'),'r').each do |object|
          @templates = YAML::load(object)
        end
      end
      @templates ||= {}
    end


    def save_state
      Dir.chdir(@wdir) do
        File.open(@dbfile,'w') do |f|
          f.puts YAML::dump(@templates)
        end
      end
    end


  end
end
