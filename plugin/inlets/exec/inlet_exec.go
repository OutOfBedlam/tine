package exec

import (
	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "exec",
		Factory: ExecInlet,
	})
}

func ExecInlet(ctx *engine.Context) engine.Inlet {
	conf := ctx.Config()
	commands := conf.GetStringArray("commands", []string{})
	environments := conf.GetStringArray("environments", []string{})
	namePrefix := conf.GetString("prefix", "")
	interval := conf.GetDuration("batch_interval", 0)
	if interval > 0 {
		// batch mode
		ret := &execPull{ctx: ctx, interval: interval}
		ret.commands = commands
		ret.environments = environments
		ret.namePrefix = namePrefix
		ret.ignoreError = conf.GetBool("ignore_error", false)
		ret.timeout = conf.GetDuration("timeout", interval)
		ret.countLimit = int64(conf.GetInt("count", 0))
		ret.trimSpace = conf.GetBool("trim_space", false)
		return ret
	} else {
		// stream mode
		ret := &execPush{ctx: ctx}
		ret.commands = commands
		ret.environments = environments
		ret.namePrefix = namePrefix
		ret.separator = []byte(conf.GetString("separator", ""))
		ret.trimSpace = conf.GetBool("trim_space", false)
		return ret
	}
}
