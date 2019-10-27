# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

Fixed:

* http: http.Client follows redirects by default. Disable this behavior to be able to test redirects.

## [v0.1.1] - Oct 27, 2019

Fixed:

* asset files' build URLs haven't contained archive name. v0.1.2 version of julian7/sensulib fixed this issue

## [v0.1.0] - Oct 27, 2019

Added:

* Initial release
* filesystem check
* http check

[Unreleased]: https://github.com/julian7/sensu-base-checks
[v0.1.0]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.1.0
[v0.1.1]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.1.1
