package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/julian7/sensu-base-checks/metrics"
	"github.com/julian7/sensulib"
	"github.com/shirou/gopsutil/disk"
	"github.com/spf13/cobra"
)

type filesystemMetricsConfig struct {
	fs  *filesystemConfig
	log *metrics.Metrics
}

func filesystemMetricsCmd() *cobra.Command {
	config := &filesystemMetricsConfig{fs: &filesystemConfig{}}
	cmd := sensulib.NewCommand(
		config,
		"filesystem",
		"Local filesystem metrics",
		"Shows free/size measurements for locally mounted filesystems.",
	)
	flags := cmd.Flags()
	config.fs.setFlags(flags)

	return cmd
}

func (conf *filesystemMetricsConfig) Run(cmd *cobra.Command, args []string) error {
	err := conf.fs.check()
	if err != nil {
		return sensulib.Unknown(err)
	}

	conf.log = metrics.New("filesystem")

	errs := sensulib.NewErrors()

	if err := conf.fs.forEach(func(part *disk.PartitionStat) {
		errs.Add(conf.checkPartition(part))
	}); err != nil {
		return err
	}

	return errs.Return(nil)
}

func (conf *filesystemMetricsConfig) checkPartition(part *disk.PartitionStat) *sensulib.Error {
	st, err := disk.Usage(part.Mountpoint)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil
		}

		return sensulib.Warn(fmt.Errorf("unable to read %s: %v", part.Mountpoint, err))
	}

	log := conf.log.With(map[string]string{"partition": part.Mountpoint})

	log.Log("bytes.free", st.Free)
	log.Log("bytes.total", st.Total)

	if st.InodesTotal > 0 {
		log.Log("inodes.free", st.InodesFree)
		log.Log("inodes.total", st.InodesTotal)
	}

	return nil
}
