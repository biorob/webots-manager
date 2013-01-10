require 'configatron'
require 'rbconfig'



case RbConfig::CONFIG['host_os']
when /linux|cygwin/
  configatron.webots_home = ENV['WEBOTS_HOME'] || '/usr/local/webots'
  configatron.install_prefix = '/usr/local/webots_manager'
when /mac|darwin/
  configatron.webots_home = ENV['WEBOTS_HOME'] || '/Applications/Webots'
  configatron.install_prefix = '/usr/local/webots_manager'
when /mswin|win|mingw/
  configatron.webots_home = ENV['WEBOTS_HOME'] || 'C:/Progam Files/Webots'
  configatron.install_prefix = 'C:/Program Files/webots_manager'
end

configatron.webots_prefix = configatron.webots_home.sub(/\/[Ww]ebots\/*\Z/,'')
