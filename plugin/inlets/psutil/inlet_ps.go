package psutil

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"slices"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "cpu",
		Factory: CpuInlet,
	})
	engine.RegisterInlet(&engine.InletReg{
		Name:    "load",
		Factory: LoadInlet,
	})

	engine.RegisterInlet(&engine.InletReg{
		Name:    "mem",
		Factory: MemInlet,
	})

	engine.RegisterInlet(&engine.InletReg{
		Name:    "disk",
		Factory: DiskInlet,
	})

	engine.RegisterInlet(&engine.InletReg{
		Name:    "diskio",
		Factory: DiskioInlet,
	})

	engine.RegisterInlet(&engine.InletReg{
		Name:    "net",
		Factory: NetInlet,
	})

	if runtime.GOOS != "darwin" {
		// darwin does not support netstat
		engine.RegisterInlet(&engine.InletReg{
			Name:    "netstat",
			Factory: NetstatInlet,
		})
	}

	engine.RegisterInlet(&engine.InletReg{
		Name:    "sensors",
		Factory: SensorsInlet,
	})

	engine.RegisterInlet(&engine.InletReg{
		Name:    "host",
		Factory: HostInlet,
	})
}

var defaultInterval = 10 * time.Second

func CpuInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	perCpu := conf.GetBool("percpu", false)
	totalCpu := conf.GetBool("totalcpu", true)
	interval := conf.GetDuration("interval", defaultInterval)
	count := conf.GetInt64("count", 0)
	return engine.InletWithPullFunc(func() ([]engine.Record, error) {
		tv, err := cpu.Percent(0, false)
		if err != nil {
			fmt.Println("ERR", err)
			return nil, err
		}
		rec := engine.NewRecord()
		if tv != nil {
			if totalCpu {
				rec = rec.Append(
					engine.NewFloatField("total_percent", tv[0]),
				)
			}
		}
		if perCpu {
			v, err := cpu.Percent(0, true)
			if err != nil {
				return []engine.Record{rec}, err
			}
			for i, p := range v {
				rec = rec.Append(
					engine.NewFloatField(fmt.Sprintf("%d_percent", i), p),
				)
			}
		}
		return []engine.Record{rec}, nil
	}, engine.WithInterval(interval), engine.WithRunCountLimit(count))
}

func LoadInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	loads := conf.GetIntSlice("loads", []int{1, 5, 15})
	interval := conf.GetDuration("interval", defaultInterval)
	count := conf.GetInt64("count", 0)
	return engine.InletWithPullFunc(func() ([]engine.Record, error) {
		stat, err := load.Avg()
		if err != nil {
			return nil, err
		}
		rec := engine.NewRecord()
		for _, i := range loads {
			switch i {
			case 1:
				rec = rec.Append(engine.NewFloatField("load1", stat.Load1))
			case 5:
				rec = rec.Append(engine.NewFloatField("load5", stat.Load5))
			case 15:
				rec = rec.Append(engine.NewFloatField("load15", stat.Load15))
			}
		}
		return []engine.Record{rec}, nil
	}, engine.WithInterval(interval), engine.WithRunCountLimit(count))
}

func MemInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	interval := conf.GetDuration("interval", defaultInterval)
	count := conf.GetInt64("count", 0)
	return engine.InletWithPullFunc(func() ([]engine.Record, error) {
		stat, err := mem.VirtualMemory()
		if err != nil {
			return nil, err
		}
		rec := engine.NewRecord(
			engine.NewUintField("total", stat.Total),
			engine.NewUintField("free", stat.Free),
			engine.NewUintField("used", stat.Used),
			engine.NewFloatField("used_percent", stat.UsedPercent),
		)
		return []engine.Record{rec}, nil
	}, engine.WithInterval(interval), engine.WithRunCountLimit(count))
}

func DiskInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	mountpoints := conf.GetStringSlice("mount_points", []string{})
	ignorefs := conf.GetStringSlice("ignore_fs", []string{})
	interval := conf.GetDuration("interval", defaultInterval)
	count := conf.GetInt64("count", 0)
	return engine.InletWithPullFunc(func() ([]engine.Record, error) {
		stat, err := disk.Partitions(false)
		if err != nil {
			return nil, err
		}
		ret := []engine.Record{}
		for _, v := range stat {
			if slices.Contains(ignorefs, v.Fstype) {
				continue
			}
			matched := len(mountpoints) == 0
			if !matched {
				for _, point := range mountpoints {
					if point == v.Mountpoint {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
			usage, err := disk.Usage(v.Mountpoint)
			if err != nil {
				return nil, err
			}
			rec := engine.NewRecord(
				engine.NewStringField("mount_point", v.Mountpoint),
				engine.NewStringField("device", v.Device),
				engine.NewStringField("fstype", v.Fstype),
				engine.NewUintField("total", usage.Total),
				engine.NewUintField("free", usage.Free),
				engine.NewUintField("used", usage.Used),
				engine.NewFloatField("used_percent", usage.UsedPercent),
			)

			if runtime.GOOS != "windows" {
				rec = rec.Append(
					engine.NewUintField("inodes_total", usage.InodesTotal),
					engine.NewUintField("inodes_free", usage.InodesFree),
					engine.NewUintField("inodes_used", usage.InodesUsed),
					engine.NewFloatField("inodes_used_percent", usage.InodesUsedPercent),
				)
			}
			ret = append(ret, rec)
		}
		return ret, nil
	}, engine.WithInterval(interval), engine.WithRunCountLimit(count))
}

func DiskioInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	devPatterns := conf.GetStringSlice("devices", []string{})
	interval := conf.GetDuration("interval", defaultInterval)
	count := conf.GetInt64("count", 0)
	return engine.InletWithPullFunc(func() ([]engine.Record, error) {
		stat, err := disk.IOCounters()
		if err != nil {
			return nil, err
		}
		ret := []engine.Record{}
		for _, v := range stat {
			matched := len(devPatterns) == 0
			for _, pattern := range devPatterns {
				if ok, err := filepath.Match(pattern, v.Name); !ok || err != nil {
					continue
				} else {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
			rec := engine.NewRecord(
				engine.NewStringField("device", v.Name),
				engine.NewStringField("serial_number", v.SerialNumber),
				engine.NewStringField("label", v.Label),
				engine.NewUintField("read_count", v.ReadCount),
				engine.NewUintField("merged_read_count", v.MergedReadCount),
				engine.NewUintField("write_count", v.WriteCount),
				engine.NewUintField("merged_write_count", v.MergedWriteCount),
				engine.NewUintField("read_bytes", v.ReadBytes),
				engine.NewUintField("write_bytes", v.WriteBytes),
				engine.NewUintField("read_time", v.ReadTime),
				engine.NewUintField("write_time", v.WriteTime),
				engine.NewUintField("iops_in_progress", v.IopsInProgress),
				engine.NewUintField("io_time", v.IoTime),
				engine.NewUintField("weighted_io", v.WeightedIO),
			)
			ret = append(ret, rec)
		}
		return ret, nil
	}, engine.WithInterval(interval), engine.WithRunCountLimit(count))
}

func NetInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	nicPatterns := conf.GetStringSlice("devices", []string{"*"})
	interval := conf.GetDuration("interval", defaultInterval)
	count := conf.GetInt64("count", 0)
	return engine.InletWithPullFunc(func() ([]engine.Record, error) {
		stat, err := net.IOCounters(true)
		if err != nil {
			return nil, err
		}
		ret := []engine.Record{}
		for _, v := range stat {
			matched := len(nicPatterns) == 0
			for _, pattern := range nicPatterns {
				if ok, err := filepath.Match(pattern, v.Name); !ok || err != nil {
					continue
				} else {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
			rec := engine.NewRecord(
				engine.NewStringField("device", v.Name),
				engine.NewUintField("bytes_sent", v.BytesSent),
				engine.NewUintField("bytes_recv", v.BytesRecv),
				engine.NewUintField("packets_sent", v.PacketsSent),
				engine.NewUintField("packets_recv", v.PacketsRecv),
				engine.NewUintField("err_in", v.Errin),
				engine.NewUintField("err_out", v.Errout),
				engine.NewUintField("drop_in", v.Dropin),
				engine.NewUintField("drop_out", v.Dropout),
				engine.NewUintField("fifos_in", v.Fifoin),
				engine.NewUintField("fifos_out", v.Fifoout),
			)
			ret = append(ret, rec)
		}
		return ret, nil
	}, engine.WithInterval(interval), engine.WithRunCountLimit(count))
}

func NetstatInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	protos := conf.GetStringSlice("protocols", []string{})
	interval := conf.GetDuration("interval", defaultInterval)
	count := conf.GetInt64("count", 0)
	return engine.InletWithPullFunc(func() ([]engine.Record, error) {
		stat, err := net.ProtoCounters(protos)
		if err != nil {
			return nil, err
		}
		ret := []engine.Record{}
		for _, st := range stat {
			rec := engine.NewRecord(
				engine.NewStringField("protocol", st.Protocol),
			)
			for key, val := range st.Stats {
				rec = rec.Append(engine.NewIntField(camelToSnake(key), val))
			}
			ret = append(ret, rec)
		}
		return ret, nil
	}, engine.WithInterval(interval), engine.WithRunCountLimit(count))
}

func camelToSnake(s string) string {
	ret := ""
	for i, c := range s {
		if 'A' <= c && c <= 'Z' {
			if i > 0 {
				ret += "_"
			}
			ret += string(c + 32)
		} else {
			ret += string(c)
		}
	}
	return ret
}

func SensorsInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	interval := conf.GetDuration("interval", defaultInterval)
	count := conf.GetInt64("count", 0)
	return engine.InletWithPullFunc(func() ([]engine.Record, error) {
		stat, err := sensors.SensorsTemperatures()
		if err != nil {
			return nil, err
		}

		ret := []engine.Record{}
		for _, v := range stat {
			rec := engine.NewRecord(
				engine.NewStringField("sensor_key", v.SensorKey),
				engine.NewFloatField("temperature", v.Temperature),
				engine.NewFloatField("high", v.High),
				engine.NewFloatField("critical", v.Critical),
			)
			ret = append(ret, rec)
		}
		return ret, nil
	}, engine.WithInterval(interval), engine.WithRunCountLimit(count))
}

func HostInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	interval := conf.GetDuration("interval", defaultInterval)
	count := conf.GetInt64("count", 0)
	return engine.InletWithPullFunc(func() ([]engine.Record, error) {
		stat, err := host.Info()
		if err != nil {
			return nil, err
		}
		rec := engine.NewRecord(
			engine.NewStringField("hostname", stat.Hostname),
			engine.NewUintField("uptime", stat.Uptime),
			engine.NewUintField("procs", stat.Procs),
			engine.NewStringField("os", stat.OS),
			engine.NewStringField("platform", stat.Platform),
			engine.NewStringField("platform_family", stat.PlatformFamily),
			engine.NewStringField("platform_version", stat.PlatformVersion),
			engine.NewStringField("kernel_version", stat.KernelVersion),
			engine.NewStringField("kernel_arch", stat.KernelArch),
			engine.NewStringField("virtualization_system", stat.VirtualizationSystem),
			engine.NewStringField("virtualization_role", stat.VirtualizationRole),
			engine.NewStringField("host_id", stat.HostID),
		)
		return []engine.Record{rec}, nil
	}, engine.WithInterval(interval), engine.WithRunCountLimit(count))
}
