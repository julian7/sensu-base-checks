package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/beevik/ntp"
	"github.com/julian7/sensu-base-checks/metrics"
	"github.com/julian7/sensulib"
	"github.com/spf13/cobra"
)

type timeConfig struct {
	server  string
	warnS   string
	warn    time.Duration
	critS   string
	crit    time.Duration
	metrics bool
}

func timeCmd() *cobra.Command {
	config := &timeConfig{}
	cmd := sensulib.NewCommand(
		config,
		"time",
		"Time drift check",
		"Measures and warns on system clock time drifts.",
	)
	flags := cmd.Flags()
	flags.StringVarP(&config.server, "server", "s", "pool.ntp.org", "NTP server used for drift detection")
	flags.StringVarP(&config.warnS, "warn", "w", "1s", "Warn on drift higher than this duration")
	flags.StringVarP(&config.critS, "crit", "c", "5s", "Crit on drift higher than this duration")
	flags.BoolVarP(&config.metrics, "metrics", "m", false, "Output measurements in TSDB format")

	return cmd
}

func (conf *timeConfig) check() error {
	var err error

	for _, item := range []struct {
		name   string
		source string
		target *time.Duration
	}{
		{"warn", conf.warnS, &conf.warn},
		{"crit", conf.critS, &conf.crit},
	} {
		*item.target, err = time.ParseDuration(item.source)
		if err != nil {
			return fmt.Errorf("parsing --%s: %w", item.name, err)
		}
	}

	for _, item := range []struct {
		name        string
		requirement bool
	}{
		{"--crit should be set", conf.crit > 0},
		{"--warn should be set", conf.warn > 0},
		{"--crit should be higher than --warn", conf.crit > conf.warn},
	} {
		if !item.requirement {
			return errors.New(item.name)
		}
	}

	return nil
}

func (conf *timeConfig) Run(cmd *cobra.Command, args []string) error {
	if err := conf.check(); err != nil {
		return err
	}

	wallclock := time.Now()

	resp, err := ntp.Query(conf.server)
	if err != nil {
		return sensulib.Warn(err)
	}

	if err := resp.Validate(); err != nil {
		return sensulib.Warn(err)
	}

	drift := resp.Time.Sub(wallclock)

	if conf.metrics {
		conf.print(drift)
		return nil
	}

	if abs(drift) > conf.crit {
		return sensulib.Crit(fmt.Errorf("clock skew is %s, level is %s", drift, conf.crit))
	}

	if abs(drift) > conf.warn {
		return sensulib.Warn(fmt.Errorf("clock skew is %s", drift))
	}

	return sensulib.Ok(errors.New("clock is adequately set"))
}

func (conf *timeConfig) print(drift time.Duration) {
	metrics.New("time").Log("ntp.offset", drift.Microseconds())
}

func abs(n time.Duration) time.Duration {
	y := n >> 63
	return (n ^ y) - y
}
