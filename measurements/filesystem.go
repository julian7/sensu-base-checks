package measurements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/julian7/sensulib"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/spf13/pflag"
)

type Filesystem struct {
	inctype  []string
	exctype  []string
	incmnt   []string
	excmnt   []string
	excopt   []string
	excpathS string
	excpath  *regexp.Regexp
}

func (conf *Filesystem) SetFlags(flags *pflag.FlagSet) {
	flags.StringSliceVarP(&conf.inctype, "inctype", "t", nil, "Filter for filesystem types")
	flags.StringSliceVarP(&conf.exctype, "exctype", "T", nil, "Ignore filesystem types")
	flags.StringSliceVarP(&conf.incmnt, "incmnt", "m", nil, "Include mount points")
	flags.StringSliceVarP(&conf.excmnt, "excmnt", "M", nil, "Ignore mount points")
	flags.StringVarP(&conf.excpathS, "excpath", "p", "", "Ignore path regular expression")
	flags.StringSliceVarP(&conf.excopt, "excopt", "o", nil, "Ignore options")
}

func (conf *Filesystem) Check() error {
	var err error

	if len(conf.excpathS) > 0 {
		conf.excpath, err = regexp.Compile(conf.excpathS)
		if err != nil {
			return fmt.Errorf("cannot interpret regexp from --excpath: %w", err)
		}
	}

	return nil
}

func (conf *Filesystem) ForEach(cb func(*disk.PartitionStat)) error {
	parts, err := disk.Partitions(true)
	if err != nil {
		return sensulib.Unknown(fmt.Errorf("cannot read partitions: %w", err))
	}

	for _, part := range parts {
		part := part

		included := includes(part.Fstype, conf.inctype) ||
			includes(part.Mountpoint, conf.incmnt)
		excluded := includes(part.Fstype, conf.exctype) ||
			includes(part.Mountpoint, conf.excmnt) ||
			hasOpt(conf.excopt, part.Opts) ||
			matchesPath(conf.excpath, part.Mountpoint) ||
			!directDevice(part.Device)

		if !excluded || included {
			cb(&part)
		}
	}

	return nil
}

func includes(needle string, haystack []string) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, accepted := range haystack {
		if accepted == needle {
			return true
		}
	}

	return false
}

func hasOpt(needles []string, haystack []string) bool {
	if len(needles) == 0 {
		return false
	}

	if len(haystack) == 0 {
		return false
	}

	for _, hay := range haystack {
		for _, needle := range needles {
			if hay == needle {
				return true
			}
		}
	}

	return false
}

func matchesPath(re *regexp.Regexp, mountpoint string) bool {
	return re != nil && re.Match([]byte(mountpoint))
}

func directDevice(device string) bool {
	return strings.Contains(device, "/") || strings.Contains(device, ":")
}
