# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

Added:

* json key/val equality check, using JMESPath

## [v0.1.4] - Dec 13, 2019

Added:

* official builds: "official" target to mage
* .bonsai.yml for [bonsai](https://bonsai.sensu.io/) registration

Changed:

* upgrade to sensulib v0.2.1
* filesystem: swallow permission denied errors

## [v0.1.3] - Nov 17, 2019

Added:

* New mage target: release (runs goshipdone, while enabling publish modules)

Changed:

* filesystem: swallow permission denied errors

Fixed:

* Updated to goshipdone v0.3.0

## [v0.1.2] - Nov 17, 2019

Added:

* Upload assets

Fixed:

* http: return with OK value if no issues happened.
* http: http.Client follows redirects by default. Disable this behavior to be able to test redirects.
* http: if expecting a 3xx response code, a redirect is by no means unexpected.

## [v0.1.1] - Oct 27, 2019

Fixed:

* asset files' build URLs haven't contained archive name. v0.1.2 version of julian7/sensulib fixed this issue

## [v0.1.0] - Oct 27, 2019

Added:

* Initial release
* filesystem check
* http check

[Unreleased]: https://github.com/julian7/sensu-base-checks
[v0.1.4]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.1.4
[v0.1.3]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.1.3
[v0.1.2]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.1.2
[v0.1.1]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.1.1
[v0.1.0]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.1.0
