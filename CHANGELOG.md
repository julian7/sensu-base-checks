# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

Added:

- ARMv7 and ARM64 Linux support (#4)

Changed:

* Updated dependencies, update go to 1.17

## [v0.4.0] - Aug 19, 2021

Fixed:

* Try all available NTP servers until one answers. It can fix intermittent NTP issues querying only one server.
* Deep certificate expiration check
* Allow certificate expiration to be checked while skipping cert verification [#2]

## [v0.3.1] - Dec 23, 2020

Changed:

* Updated dependencies

Fixed:

* rebuilt with go 1.15 to avoid [mlock error](https://github.com/golang/go/issues/37436)

## [v0.3.0] - Mar 20, 2020

Added:

* http: new measurements: body size, body transfer speed, body transfer errors

## [v0.2.2] - Feb 10, 2020

Fixed:

* filesystem: fix crash when showing zero bytes

## [v0.2.1] - Jan 26, 2020

Changed:

* filesystem: filtering out synthetic filesystems are implemented as an exclude check,
  and now it allows "/" anywhere in the device name (making ZFS subvolumes visible; only
  main volumes have to be included manually)

Fixed:

* time: use SNTP's clock offset instead of computing from wallclock
* time: opentsdb requires at least one tag

## [v0.2.0] - Jan 19, 2020

Added:

* http: json key/val equality check, using JMESPath
* time: NTP check / metrics

Changed:

* filesystem: collect metrics for mounted filesystems
* http: CURL-like metrics

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
[v0.4.0]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.4.0
[v0.3.1]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.3.1
[v0.3.0]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.3.0
[v0.2.2]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.2.2
[v0.2.1]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.2.1
[v0.2.0]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.2.0
[v0.1.4]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.1.4
[v0.1.3]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.1.3
[v0.1.2]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.1.2
[v0.1.1]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.1.1
[v0.1.0]: https://github.com/julian7/sensu-base-checks/releases/tag/v0.1.0
