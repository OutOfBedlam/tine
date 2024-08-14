package psutil

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"slices"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/OutOfBedlam/tine/util"
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
	return engine.InletWithFunc(func() ([]engine.Record, error) {
		tv, err := cpu.Percent(0, false)
		if err != nil {
			fmt.Println("ERR", err)
			return nil, err
		}
		rec := engine.NewRecord()
		if tv != nil {
			if totalCpu {
				rec = rec.Append(
					engine.NewField("total_percent", tv[0]),
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
					engine.NewField(fmt.Sprintf("%d_percent", i), p),
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
	return engine.InletWithFunc(func() ([]engine.Record, error) {
		stat, err := load.Avg()
		if err != nil {
			return nil, err
		}
		rec := engine.NewRecord()
		for _, i := range loads {
			switch i {
			case 1:
				rec = rec.Append(engine.NewField("load1", stat.Load1))
			case 5:
				rec = rec.Append(engine.NewField("load5", stat.Load5))
			case 15:
				rec = rec.Append(engine.NewField("load15", stat.Load15))
			}
		}
		return []engine.Record{rec}, nil
	}, engine.WithInterval(interval), engine.WithRunCountLimit(count))
}

func MemInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	interval := conf.GetDuration("interval", defaultInterval)
	count := conf.GetInt64("count", 0)
	return engine.InletWithFunc(func() ([]engine.Record, error) {
		stat, err := mem.VirtualMemory()
		if err != nil {
			return nil, err
		}
		rec := engine.NewRecord(
			engine.NewField("total", stat.Total),
			engine.NewField("free", stat.Free),
			engine.NewField("used", stat.Used),
			engine.NewField("used_percent", stat.UsedPercent),
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
	return engine.InletWithFunc(func() ([]engine.Record, error) {
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
				engine.NewField("mount_point", v.Mountpoint),
				engine.NewField("device", v.Device),
				engine.NewField("fstype", v.Fstype),
				engine.NewField("total", usage.Total),
				engine.NewField("free", usage.Free),
				engine.NewField("used", usage.Used),
				engine.NewField("used_percent", usage.UsedPercent),
			)

			if runtime.GOOS != "windows" {
				rec = rec.Append(
					engine.NewField("inodes_total", usage.InodesTotal),
					engine.NewField("inodes_free", usage.InodesFree),
					engine.NewField("inodes_used", usage.InodesUsed),
					engine.NewField("inodes_used_percent", usage.InodesUsedPercent),
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
	return engine.InletWithFunc(func() ([]engine.Record, error) {
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
				engine.NewField("device", v.Name),
				engine.NewField("serial_number", v.SerialNumber),
				engine.NewField("label", v.Label),
				engine.NewField("read_count", v.ReadCount),
				engine.NewField("merged_read_count", v.MergedReadCount),
				engine.NewField("write_count", v.WriteCount),
				engine.NewField("merged_write_count", v.MergedWriteCount),
				engine.NewField("read_bytes", v.ReadBytes),
				engine.NewField("write_bytes", v.WriteBytes),
				engine.NewField("read_time", v.ReadTime),
				engine.NewField("write_time", v.WriteTime),
				engine.NewField("iops_in_progress", v.IopsInProgress),
				engine.NewField("io_time", v.IoTime),
				engine.NewField("weighted_io", v.WeightedIO),
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
	return engine.InletWithFunc(func() ([]engine.Record, error) {
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
				engine.NewField("device", v.Name),
				engine.NewField("bytes_sent", v.BytesSent),
				engine.NewField("bytes_recv", v.BytesRecv),
				engine.NewField("packets_sent", v.PacketsSent),
				engine.NewField("packets_recv", v.PacketsRecv),
				engine.NewField("err_in", v.Errin),
				engine.NewField("err_out", v.Errout),
				engine.NewField("drop_in", v.Dropin),
				engine.NewField("drop_out", v.Dropout),
				engine.NewField("fifos_in", v.Fifoin),
				engine.NewField("fifos_out", v.Fifoout),
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
	return engine.InletWithFunc(func() ([]engine.Record, error) {
		stat, err := net.ProtoCounters(protos)
		if err != nil {
			return nil, err
		}
		ret := []engine.Record{}
		for _, st := range stat {
			rec := engine.NewRecord(
				engine.NewField("protocol", st.Protocol),
			)
			for key, val := range st.Stats {
				rec = rec.Append(engine.NewField(util.CamelToSnake(key), val))
			}
			ret = append(ret, rec)
		}
		return ret, nil
	}, engine.WithInterval(interval), engine.WithRunCountLimit(count))
}

func SensorsInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	interval := conf.GetDuration("interval", defaultInterval)
	count := conf.GetInt64("count", 0)
	return engine.InletWithFunc(func() ([]engine.Record, error) {
		stat, err := sensors.SensorsTemperatures()
		if err != nil {
			return nil, err
		}

		ret := []engine.Record{}
		for _, v := range stat {
			rec := engine.NewRecord(
				engine.NewField("sensor_key", v.SensorKey),
				engine.NewField("temperature", v.Temperature),
				engine.NewField("high", v.High),
				engine.NewField("critical", v.Critical),
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
	return engine.InletWithFunc(func() ([]engine.Record, error) {
		stat, err := host.Info()
		if err != nil {
			return nil, err
		}
		rec := engine.NewRecord(
			engine.NewField("hostname", stat.Hostname),
			engine.NewField("uptime", stat.Uptime),
			engine.NewField("procs", stat.Procs),
			engine.NewField("os", stat.OS),
			engine.NewField("platform", stat.Platform),
			engine.NewField("platform_family", stat.PlatformFamily),
			engine.NewField("platform_version", stat.PlatformVersion),
			engine.NewField("kernel_version", stat.KernelVersion),
			engine.NewField("kernel_arch", stat.KernelArch),
			engine.NewField("virtualization_system", stat.VirtualizationSystem),
			engine.NewField("virtualization_role", stat.VirtualizationRole),
			engine.NewField("host_id", stat.HostID),
		)
		return []engine.Record{rec}, nil
	}, engine.WithInterval(interval), engine.WithRunCountLimit(count))
}
