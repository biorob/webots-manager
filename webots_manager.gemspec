# -*- encoding: utf-8 -*-
lib = File.expand_path('../lib', __FILE__)
$LOAD_PATH.unshift(lib) unless $LOAD_PATH.include?(lib)
require 'webots_manager/version'

Gem::Specification.new do |gem|
  gem.name          = "webots_manager"
  gem.version       = WebotsManager::VERSION
  gem.authors       = ["Alexandre Tuleu"]
  gem.email         = ["alexandre.tuleu.2005@polytechnique.org"]
  gem.description   = "A small script to manage multiple webots installation."
  gem.summary       = "A small script to manage multiple webots installation. Much like rvm"
  gem.homepage      = "https://github.com/biorob/webots_manager"

  gem.add_dependency("thor")
  gem.add_dependency("configatron")

  gem.files         = `git ls-files`.split($/)
  gem.executables   = gem.files.grep(%r{^bin/}).map{ |f| File.basename(f) }
  gem.test_files    = gem.files.grep(%r{^(test|spec|features)/})
  gem.require_paths = ["lib"]
end
