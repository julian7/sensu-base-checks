# sensu-base-checks

Baseline health checks for Sensu, written in go.

Very early in development, just a couple of checks finished:

- filesystem: it is modeled after sensu-plugins-disk-checks' [check-disk-usage.rb](https://github.com/sensu-plugins/sensu-plugins-disk-checks/blob/master/bin/check-disk-usage.rb) script
- http: modeled after sensu-plugins-http's [check-http.rb](https://github.com/sensu-plugins/sensu-plugins-http/blob/master/bin/check-http.rb) script

## Goals

There are three goals for this project:

1. a direct goal: provide a set of basic tests, which can replace currently used ruby checks
2. write a go library to help writing go check plugins. This can easily spin off as a separate project, with wider goal (like supporting metrics, or other kinds of extensions for Sensu GO)
3. provide a [magefile](https://magefile.org/) to help publishing assets for Sensu GO

## Legal

This project is licensed under [Blue Oak Model License v1.0.0](https://blueoakcouncil.org/license/1.0.0). It is not registered either at OSI or GNU, therefore GitHub is widely looking at the other direction. However, this is the license I'm most happy with: you can read and understand it with no legal degree, and there are no hidden or cryptic meanings in it.

The project is also governed with [Contributor Covenant](https://contributor-covenant.org/)'s [Code of Conduct](https://www.contributor-covenant.org/version/1/4/) in mind. I'm not copying it here, as a pledge for taking the verbatim version by the word, and we are not going to modify it in any way.

## Any issues?

Open a ticket, perhaps a pull request. We support [GitHub Flow](https://guides.github.com/introduction/flow/). You might want to [fork](https://guides.github.com/activities/forking/) this project first.
