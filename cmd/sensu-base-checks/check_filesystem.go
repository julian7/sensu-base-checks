package main

import (
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/julian7/sensulib"
	"github.com/shirou/gopsutil/disk"
	"github.com/spf13/cobra"
)

type filesystemCheckConfig struct {
	fs      *filesystemConfig
	bwarn   float64
	bcrit   float64
	iwarn   float64
	icrit   float64
	magic   float64
	normal  int
	minimum int
}

func filesystemCheckCmd() *cobra.Command {
	config := &filesystemCheckConfig{fs: &filesystemConfig{}}
	cmd := sensulib.NewCommand(
		config,
		"filesystem",
		"Local filesystem check",
		"Checks for locally mounted filesystems.",
	)
	flags := cmd.Flags()
	config.fs.setFlags(flags)
	flags.Float64VarP(&config.bwarn, "bwarn", "w", 85.0, "Warn if PERCENT or more of filesystem full; (0,100]")
	flags.Float64VarP(&config.bcrit, "bcrit", "c", 95.0, "Critical if PERCENT or more of filesystem full; (0,100]")
	flags.Float64VarP(&config.iwarn, "iwarn", "W", 85.0, "Warn if PERCENT or more of inodes used; (0,100]")
	flags.Float64VarP(&config.icrit, "icrit", "C", 95.0, "Critical if PERCENT or more of inodes used; (0,100]")
	flags.Float64VarP(&config.magic, "magic", "x", 1.0, "Magic factor to adjust warn/crit thresholds; (0,1]")
	flags.IntVarP(&config.normal, "normal", "n", 20, "Levels are not adapted for filesystems of exactly this size (GB)."+
		" Levels reduced below this size, and raised for larger sizes.")
	flags.IntVarP(&config.minimum, "minimum", "l", 100, "Minimum size to adjust (ing GB)")

	return cmd
}

func (conf *filesystemCheckConfig) check() error {
	if err := conf.fs.check(); err != nil {
		return err
	}

	checks := []struct {
		name   string
		check  bool
		errstr string
	}{
		{"bwarn", conf.bwarn < 0, "higher than 0"},
		{"bwarn", conf.bwarn > 100, "at most 100"},
		{"bcrit", conf.bcrit < 0, "higher than 0"},
		{"bcrit", conf.bcrit > 100, "at most 100"},
		{"bcrit", conf.bcrit <= conf.bwarn, "should be higher than --bwarn"},
		{"iwarn", conf.iwarn < 0, "higher than 0"},
		{"iwarn", conf.iwarn > 100, "at most 100"},
		{"icrit", conf.icrit < 0, "higher than 0"},
		{"icrit", conf.icrit > 100, "at most 100"},
		{"icrit", conf.icrit <= conf.iwarn, "should be higher than --iwarn"},
		{"magic", conf.magic < 0, "higher than 0"},
		{"magic", conf.magic > 1, "at most 1"},
	}

	for _, check := range checks {
		if check.check {
			return fmt.Errorf("--%s should be %s", check.name, check.errstr)
		}
	}

	return nil
}

func (conf *filesystemCheckConfig) Run(cmd *cobra.Command, args []string) error {
	err := conf.check()
	if err != nil {
		return sensulib.Unknown(err)
	}

	errs := sensulib.NewErrors()

	if err := conf.fs.forEach(func(part *disk.PartitionStat) {
		errs.Add(conf.checkPartition(part))
	}); err != nil {
		return err
	}

	return errs.Return(sensulib.Ok(
		fmt.Errorf(
			"all filesystems are under %s storage and %s inode usage",
			sensulib.PercentToHuman(conf.bwarn, 1),
			sensulib.PercentToHuman(conf.iwarn, 1),
		),
	))
}

func adjustLevel(total, normal uint64, magic, percent float64) float64 {
	return 100 - ((100 - percent) * math.Pow(float64(total/normal), magic-1))
}

func (conf *filesystemCheckConfig) checkPartition(part *disk.PartitionStat) *sensulib.Error {
	st, err := disk.Usage(part.Mountpoint)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil
		}

		return sensulib.Warn(fmt.Errorf("unable to read %s: %v", part.Mountpoint, err))
	}

	if st.InodesTotal > 0 {
		if st.InodesUsedPercent >= conf.iwarn {
			err := fmt.Errorf(
				"%s %s inode usage",
				part.Mountpoint,
				sensulib.PercentToHuman(st.InodesUsedPercent, 1),
			)

			if st.InodesUsedPercent >= conf.icrit {
				return sensulib.Crit(err)
			}

			return sensulib.Warn(err)
		}
	}

	var bcrit, bwarn float64

	normal := uint64(conf.normal) * 1024 * 1024
	minimum := uint64(conf.minimum) * 1024 * 1024

	if st.Total <= minimum {
		bwarn = conf.bwarn
		bcrit = conf.bcrit
	} else {
		bwarn = adjustLevel(st.Total, normal, conf.magic, conf.bwarn)
		bcrit = adjustLevel(st.Total, normal, conf.magic, conf.bcrit)
	}

	if st.UsedPercent >= bwarn {
		err = fmt.Errorf(
			"%s %s usage (%s free of %s)",
			part.Mountpoint,
			sensulib.PercentToHuman(st.UsedPercent, 2),
			sensulib.SizeToHuman(st.Free),
			sensulib.SizeToHuman(st.Total),
		)

		if st.UsedPercent >= bcrit {
			return sensulib.Crit(err)
		}

		return sensulib.Warn(err)
	}

	return nil
}
