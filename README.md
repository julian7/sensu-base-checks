# sensu-base-checks

Baseline health and metric checks for Sensu, written in go.

## Subcommands

Almost all subcommands support the `--metrics` option (there is no short form to it), which suppresses health checks, and emits measurements in [OpenTSDB](http://opentsdb.net/) format.

### filesystem

This check is modeled after sensu-plugins-disk-checks' [check-disk-usage.rb](https://github.com/sensu-plugins/sensu-plugins-disk-checks/blob/master/bin/check-disk-usage.rb) script.

```text
Usage:
  sensu-base-checks filesystem [flags]

Flags:
  -c, --bcrit float       Critical if PERCENT or more of filesystem full; (0,100] (default 95)
  -w, --bwarn float       Warn if PERCENT or more of filesystem full; (0,100] (default 85)
  -M, --excmnt strings    Ignore mount points
  -o, --excopt strings    Ignore options
  -p, --excpath string    Ignore path regular expression
  -T, --exctype strings   Ignore filesystem types
  -h, --help              help for filesystem
  -C, --icrit float       Critical if PERCENT or more of inodes used; (0,100] (default 95)
  -m, --incmnt strings    Include mount points
  -t, --inctype strings   Filter for filesystem types
  -W, --iwarn float       Warn if PERCENT or more of inodes used; (0,100] (default 85)
  -x, --magic float       Magic factor to adjust warn/crit thresholds; (0,1] (default 1)
      --metrics           Output measurements in OpenTSDB format
  -l, --minimum int       Minimum size to adjust (ing GB) (default 100)
  -n, --normal int        Levels are not adapted for filesystems of exactly this size (GB). Levels reduced below this size, and raised for larger sizes. (default 20)
  ```

It filters filesystems, in a way that it enumerates all not explicitly excluded or explicitly included ones. In practice, it means all inclusion options are affecting as a veto for exclusion options.

This command goes through all the selected filesystems, and enumerates free space / inode size (unix only) on them, comparing to a common percentage (free/size). For large filesystems, percentage calculation can be distorted by `magic`, `minimum`, and `normal` options using the following expression, when the filesystem size is larger than `minimum` filesystem size:

```text
100 - (100 - percent) * (size/normal)^(magic-1)
```

Examples for 95% full filesystems, with 20G normal size:

FS size | magic | distorted percentage
-----: | :---: | ----:
1 TB   | 0.95  | 95.892
1 TB   | 0.9   | 96.625
1 TB   | 0.5   | 99.3
10 TB  | 0.95  | 96.34
10 TB  | 0.9   | 97.32
10 TB  | 0.5   | 99.779

The command aggregates all the errors, showing all warning / critical level alerts, and it returns with the highest criticality issue it encountered.

When `--metrics` is provided, it returns

- filesystem.bytes.free: free bytes
- filesystem.bytes.total: total bytes
- inodes.free: free inodes (unix only)
- inodes.total: total inodes (unix only)

Tags:

- dev: source device
- fstype: filesystem type
- partition: mount point

Known issues:

- root ZFS volume is filtered, as they don't have "/" in their device names. They can be included manually though.

### http

This command has been modeled after sensu-plugins-http's [check-http.rb](https://github.com/sensu-plugins/sensu-plugins-http/blob/master/bin/check-http.rb) and [check-http-json.rb](https://github.com/sensu-plugins/sensu-plugins-http/blob/master/bin/check-http-json.rb) scripts.

This check runs a HTTP query, and inspects return values. Returns

- Unknown on configuration issues,
- Warning on nearing TLS cert expiry or not matching, but non-error HTTP codes,
- Critical on any other cases.

Timeout duration can be provided in short range (eg. ms, s, m, h), cert expiry
can be provided with longer range too (like d, w, mo).

```text
Usage:
  sensu-base-checks http [flags]

Flags:
  -d, --body string         HTTP body
  -C, --ca string           CA Certificate file
  -c, --cert string         Certificate file
  -e, --expiry string       Warn EXPIRY before cert expires (duration, like 5d)
  -H, --header strings      HTTP header
  -h, --help                help for http
  -k, --insecure            Enable insecure connections
  -K, --json-key string     JSON key selector in JMESPath syntax
  -V, --json-val string     expected value for JSON key in string form
  -X, --method string       HTTP method (default "GET")
      --metrics             Output measurements in OpenTSDB format
  -R, --redirect string     Expect redirection to
  -r, --response uint       HTTP error code to expect; use 3-digits for exact, 1-digit for first digit check (default 2)
  -t, --timeout string      Connection timeout (default "5s")
  -u, --url string          Target URL (default "http://127.0.0.1:80/")
  -A, --user-agent string   User agent
```

This command checks for:

- HTTP request timeout
- TLS certificate expiration date
- HTTP response (can be provided either in three digits, or in just the first digit)
- Redirect location match
- if the returned body is in JSON, then it can search for a single JSON key, checking whether it contains a certain value

The JSON check uses [JMESPath](http://jmespath.org/) to identify the key, and it converts the value to string using Go's [default format (%v)](https://golang.org/pkg/fmt/).

When `--metrics` is provided, it shows the following measurements:

- http.time.total: total retrieval time (in microseconds)
- http.time.namelookup: DNS resolve time (in microseconds)
- http.time.connect: time to connect (from start; in microseconds)
- http.time.pretransfer: time to TLS handshake (from start; in microseconds)
- http.time.starttransfer: time to first byte arrived (from start; in microseconds)
- http.time.body_transfer: time from first byte to finish (in microseconds)
- http.http.http_code: returned status code
- http.body_bytes: number of received bytes in HTTP body
- http.http.error: received error while reading HTTP body (`<nil>` if no error received)
- http.speed.body_transfer: body transfer speed (in bytes/s; only if no errors, and non-zero body_transfer and body_bytes values)

Provided tags:

- url: remote URL

### time

This command checks for system time to be in operation limits, or it provides this data as metrics.

```text
Measures and warns on system clock time drifts.

Usage:
  sensu-base-checks time [flags]

Flags:
  -c, --crit string     Crit on drift higher than this duration (default "5s")
  -h, --help            help for time
      --metrics         Output measurements in TSDB format
  -s, --server string   NTP server used for drift detection (default "pool.ntp.org")
  -w, --warn string     Warn on drift higher than this duration (default "1s")
```

When `--metrics` is provided, it returns a single value as `time.ntp.offset`, in microseconds.

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
