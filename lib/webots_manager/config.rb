require 'configatron'
require 'rbconfig'

configatron.cyberbotics_archive = 'http://www.cyberbotics.com/archive/'

case RbConfig::CONFIG['host_os']
when /linux|cygwin/
  configatron.webots_home         = ENV['WEBOTS_HOME'] || '/usr/local/webots'
  configatron.install_prefix      = '/usr/local/webots_manager'
  configatron.cyberbotics_archive =   configatron.cyberbotics_archive + 'linux/'
  configatron.suffix              = '.tar.bz2'
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

configatron.webots_prefix = configatron.webots_home.sub(/\/[Ww]ebots\/*\Z/,'')
