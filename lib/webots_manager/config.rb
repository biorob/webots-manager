require 'configatron'
require 'rbconfig'

configatron.cyberbotics_archive = 'http://www.cyberbotics.com/archive/'

case RbConfig::CONFIG['host_os']
when /linux|cygwin/
  configatron.webots_home         = ENV['WEBOTS_HOME'] || '/usr/local/webots'
  configatron.install_prefix      = '/usr/local/webots_manager'
  configatron.cyberbotics_archive =   configatron.cyberbotics_archive + 'linux/'
  configatron.suffix              = '.tar.bz2'
  configatron.group_name          = 'webots-manager'
# when /mac|darwin/
#   configatron.webots_home = ENV['WEBOTS_HOME'] || '/Applications/Webots'
#   configatron.install_prefix = '/usr/local/webots_manager'
#   configatron.cyberbotics_archive =   configatron.cyberbotics_archive + 'mac/'
#   configatron.suffix              = '.dmg'
# when /mswin|win|mingw/
#   configatron.webots_home = ENV['WEBOTS_HOME'] || 'C:/Progam Files/Webots'
#   configatron.install_prefix = 'C:/Program Files/webots_manager'
#   configatron.cyberbotics_archive =   configatron.cyberbotics_archive + 'windows/'
#   configatron.suffix              = '_setup.exe'
else
  raise RuntimeError, "System #{RbConfig::CONFIG['host_os']} is not supported"
end


case RbConfig::CONFIG['host_cpu']
when /i386|i486|i686/
  configatron.arch = 'i386'
when /amd64|x86[_\-]64/
  configatron.arch = 'x86-64'
end

configatron.webots_prefix = configatron.webots_home.sub(/\/[Ww]ebots\/*\Z/,'')
