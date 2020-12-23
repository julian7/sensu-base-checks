package main

import (
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/julian7/sensu-base-checks/measurements"
	"github.com/julian7/sensu-base-checks/metrics"
	"github.com/julian7/sensulib"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/spf13/cobra"
)

type filesystemConfig struct {
	fs      *measurements.Filesystem
	BWarn   float64
	BCrit   float64
	IWarn   float64
	ICrit   float64
	Magic   float64
	Metrics bool
	Minimum int
	Normal  int
	log     *metrics.Metrics
}

func filesystemCmd() *cobra.Command {
	config := &filesystemConfig{fs: &measurements.Filesystem{}}
	cmd := sensulib.NewCommand(
		config,
		"filesystem",
		"Local filesystem check",
		"Checks for locally mounted filesystems.",
	)
	flags := cmd.Flags()
	config.fs.SetFlags(flags)
	flags.Float64VarP(&config.BWarn, "bwarn", "w", 85.0, "Warn if PERCENT or more of filesystem full; (0,100]")
	flags.Float64VarP(&config.BCrit, "bcrit", "c", 95.0, "Critical if PERCENT or more of filesystem full; (0,100]")
	flags.Float64VarP(&config.IWarn, "iwarn", "W", 85.0, "Warn if PERCENT or more of inodes used; (0,100]")
	flags.Float64VarP(&config.ICrit, "icrit", "C", 95.0, "Critical if PERCENT or more of inodes used; (0,100]")
	flags.Float64VarP(&config.Magic, "magic", "x", 1.0, "Magic factor to adjust warn/crit thresholds; (0,1]")
	flags.BoolVar(&config.Metrics, "metrics", false, "Output measurements in OpenTSDB format")
	flags.IntVarP(&config.Minimum, "minimum", "l", 100, "Minimum size to adjust (ing GB)")
	flags.IntVarP(&config.Normal, "normal", "n", 20, "Levels are not adapted for filesystems of exactly this size (GB)."+
		" Levels reduced below this size, and raised for larger sizes.")

	return cmd
}

func (conf *filesystemConfig) check() error {
	if err := conf.fs.Check(); err != nil {
		return err
	}

	if conf.Metrics {
		return nil
	}

	checks := []struct {
		name   string
		check  bool
		errstr string
	}{
		{"bwarn", conf.BWarn < 0, "higher than 0"},
		{"bwarn", conf.BWarn > 100, "at most 100"},
		{"bcrit", conf.BCrit < 0, "higher than 0"},
		{"bcrit", conf.BCrit > 100, "at most 100"},
		{"bcrit", conf.BCrit <= conf.BWarn, "should be higher than --bwarn"},
		{"iwarn", conf.IWarn < 0, "higher than 0"},
		{"iwarn", conf.IWarn > 100, "at most 100"},
		{"icrit", conf.ICrit < 0, "higher than 0"},
		{"icrit", conf.ICrit > 100, "at most 100"},
		{"icrit", conf.ICrit <= conf.IWarn, "should be higher than --iwarn"},
		{"magic", conf.Magic < 0, "higher than 0"},
		{"magic", conf.Magic > 1, "at most 1"},
	}

	for _, check := range checks {
		if check.check {
			return fmt.Errorf("--%s should be %s", check.name, check.errstr)
		}
	}

	return nil
}

func (conf *filesystemConfig) Run(cmd *cobra.Command, args []string) error {
	var errDefault error

	err := conf.check()
	if err != nil {
		return sensulib.Unknown(err)
	}

	checkFn := conf.checkPartition

	if conf.Metrics {
		conf.log = metrics.New("filesystem")
		checkFn = conf.measurePartition
	} else {
		errDefault = sensulib.Ok(
			fmt.Errorf(
				"all filesystems are under %s storage and %s inode usage",
				sensulib.PercentToHuman(conf.BWarn, 1),
				sensulib.PercentToHuman(conf.IWarn, 1),
			),
		)
	}

	errs := sensulib.NewErrors()

	if err := conf.fs.ForEach(func(part *disk.PartitionStat) {
		errs.Add(checkFn(part))
	}); err != nil {
		return err
	}

	return errs.Return(errDefault)
}

func adjustLevel(total, normal uint64, magic, percent float64) float64 {
	return 100 - ((100 - percent) * math.Pow(float64(total/normal), magic-1))
}

func (conf *filesystemConfig) checkPartition(part *disk.PartitionStat) *sensulib.Error {
	st, err := disk.Usage(part.Mountpoint)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil
		}

		return sensulib.Warn(fmt.Errorf("unable to read %s: %v", part.Mountpoint, err))
	}

	if st.InodesTotal > 0 {
		if st.InodesUsedPercent >= conf.IWarn {
			err := fmt.Errorf(
				"%s %s inode usage",
				part.Mountpoint,
				sensulib.PercentToHuman(st.InodesUsedPercent, 1),
			)

			if st.InodesUsedPercent >= conf.ICrit {
				return sensulib.Crit(err)
			}

			return sensulib.Warn(err)
		}
	}

	var bcrit, bwarn float64

	normal := uint64(conf.Normal) * 1024 * 1024
	minimum := uint64(conf.Minimum) * 1024 * 1024

	if st.Total <= minimum {
		bwarn = conf.BWarn
		bcrit = conf.BCrit
	} else {
		bwarn = adjustLevel(st.Total, normal, conf.Magic, conf.BWarn)
		bcrit = adjustLevel(st.Total, normal, conf.Magic, conf.BCrit)
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

func (conf *filesystemConfig) measurePartition(part *disk.PartitionStat) *sensulib.Error {
	st, err := disk.Usage(part.Mountpoint)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil
		}

		return sensulib.Warn(fmt.Errorf("unable to read %s: %v", part.Mountpoint, err))
	}

	log := conf.log.With(map[string]string{
		"dev":       part.Device,
		"fstype":    part.Fstype,
		"partition": part.Mountpoint,
	})

	log.Log("bytes.free", st.Free)
	log.Log("bytes.total", st.Total)

	if st.InodesTotal > 0 {
		log.Log("inodes.free", st.InodesFree)
		log.Log("inodes.total", st.InodesTotal)
	}

	return nil
}
