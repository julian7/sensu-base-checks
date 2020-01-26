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
	Server  string
	WarnS   string
	warn    time.Duration
	CritS   string
	crit    time.Duration
	Metrics bool
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
	flags.StringVarP(&config.Server, "server", "s", "pool.ntp.org", "NTP server used for drift detection")
	flags.StringVarP(&config.WarnS, "warn", "w", "1s", "Warn on drift higher than this duration")
	flags.StringVarP(&config.CritS, "crit", "c", "5s", "Crit on drift higher than this duration")
	flags.BoolVar(&config.Metrics, "metrics", false, "Output measurements in OpenTSDB format")

	return cmd
}

func (conf *timeConfig) check() error {
	var err error

	for _, item := range []struct {
		name   string
		source string
		target *time.Duration
	}{
		{"warn", conf.WarnS, &conf.warn},
		{"crit", conf.CritS, &conf.crit},
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

	resp, err := ntp.Query(conf.Server)
	if err != nil {
		return sensulib.Warn(err)
	}

	if err := resp.Validate(); err != nil {
		return sensulib.Warn(err)
	}

	drift := resp.ClockOffset

	if conf.Metrics {
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
	metrics.New("time").With(map[string]string{"server": conf.Server}).Log("ntp.offset", drift.Microseconds())
}

func abs(n time.Duration) time.Duration {
	y := n >> 63
	return (n ^ y) - y
}
