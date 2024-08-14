package base

import "github.com/OutOfBedlam/tine/engine"

func init() {
	engine.RegisterFlow(&engine.FlowReg{Name: "merge", Factory: MergeFlow})
	engine.RegisterFlow(&engine.FlowReg{Name: "flatten", Factory: FlattenFlow})
	engine.RegisterFlow(&engine.FlowReg{Name: "damper", Factory: DamperFlow})
	engine.RegisterFlow(&engine.FlowReg{Name: "dump", Factory: DumpFlow})
	engine.RegisterFlow(&engine.FlowReg{Name: "select", Factory: SelectFlow})
	engine.RegisterFlow(&engine.FlowReg{Name: "update", Factory: UpdateFlow})
	engine.RegisterFlow(&engine.FlowReg{Name: "inject", Factory: InjectFlow})
}
