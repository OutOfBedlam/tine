package snmp

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "snmp",
		Factory: SnmpInlet,
	})
}

func SnmpInlet(ctx *engine.Context) engine.Inlet {
	return &snmpInlet{ctx: ctx}
}

type snmpInlet struct {
	ctx             *engine.Context
	interval        time.Duration
	runCountLimit   int
	runCount        int
	clientConf      ClientConfig
	name            string
	nameInfix       string
	agents          []string
	tables          []Table
	fields          []Field
	tags            engine.Tags
	connectionCache []Conn
	translator      Translator
}

var _ = engine.Inlet((*snmpInlet)(nil))

func (si *snmpInlet) Open() error {
	conf := si.ctx.Config()
	si.interval = conf.GetDuration("interval", 10*time.Second)
	si.runCountLimit = conf.GetInt("count", 0)
	si.name = conf.GetString("name", "inlets.snmp")
	si.nameInfix = conf.GetString("name_infix", "_")
	si.tags = engine.Tags{}

	si.clientConf.Timeout = conf.GetDuration("timeout", 5*time.Second)
	si.clientConf.Retries = conf.GetInt("retries", 3)
	si.clientConf.Version = conf.GetInt("version", 2)
	si.clientConf.UseUnconnectedUDPSocket = conf.GetBool("unconnected_udp_socket", false)
	// SNMPv1 and SNMPv2
	si.clientConf.Community = conf.GetString("community", "public")
	// SNMPv2 and SNMPv3
	si.clientConf.MaxRepetitions = conf.GetUint32("max_repetitions", 10)
	// SNMPv3
	si.clientConf.ContextName = conf.GetString("context_name", "")
	si.clientConf.SecLevel = conf.GetString("sec_level", "authNoPriv")
	si.clientConf.SecName = conf.GetString("sec_name", "username")
	si.clientConf.AuthProtocol = conf.GetString("auth_protocol", "MD5")
	si.clientConf.AuthPassword = conf.GetString("auth_password", "password")
	si.clientConf.PrivProtocol = conf.GetString("priv_protocol", "")
	si.clientConf.PrivPassword = conf.GetString("priv_password", "")
	si.clientConf.EngineID = conf.GetString("engine_id", "")
	si.clientConf.EngineBoots = conf.GetUint32("engine_boots", 0)
	si.clientConf.EngineTime = conf.GetUint32("engine_time", 0)

	si.agents = conf.GetStringSlice("agents", nil)
	if len(si.agents) == 0 {
		return fmt.Errorf("no SNMP agent specified")
	}

	mibPath := conf.GetStringSlice("mib_paths", []string{})
	mibTranslator := conf.GetString("translator", "gosmi")
	switch mibTranslator {
	case "gosmi":
		if trans, err := NewGosmiTranslator(mibPath); err != nil {
			return err
		} else {
			si.translator = trans
		}
	case "netsnmp":
		si.translator = NewNetsnmpTranslator()
	default:
		return fmt.Errorf("invalid translator %q", mibTranslator)
	}

	si.connectionCache = make([]Conn, len(si.agents))

	tables := conf.GetConfigSlice("tables", nil)
	for _, c := range tables {
		t := Table{}
		t.Name = c.GetString("name", "")
		t.Oid = c.GetString("oid", "")
		t.InheritTags = c.GetStringSlice("inherit_tags", nil)
		t.IndexAsTag = c.GetBool("index_as_tag", false)
		si.tables = append(si.tables, t)
	}
	for i := range si.tables {
		if err := (&si.tables[i]).init(si.translator); err != nil {
			return fmt.Errorf("initializing table %s: %w", si.tables[i].Name, err)
		}
	}

	fields := conf.GetConfigSlice("fields", nil)
	for _, c := range fields {
		f := Field{}
		f.Name = c.GetString("name", "")
		f.Oid = c.GetString("oid", "")
		f.OidIndexSuffix = c.GetString("oid_index_suffix", "")
		f.OidIndexLength = c.GetInt("oid_index_length", 0)
		f.IsTag = c.GetBool("is_tag", false)
		f.Conversion = c.GetString("conversion", "")
		f.Translate = c.GetBool("translate", false)
		f.SecondaryIndexTable = c.GetBool("secondary_index_table", false)
		f.SecondaryIndexUse = c.GetBool("secondary_index_use", false)
		f.SecondaryOuterJoin = c.GetBool("secondary_outer_join", false)
		si.fields = append(si.fields, f)
	}
	for i := range si.fields {
		if err := (&si.fields[i]).Init(si.translator); err != nil {
			return fmt.Errorf("initializing field %s: %w", si.fields[i].Name, err)
		}
	}

	tags := conf.GetConfigSlice("tags", nil)
	for _, c := range tags {
		name := c.GetString("name", "")
		value := c.GetValue("value")
		if value.IsNotNull() {
			si.tags[name] = value
		}
	}
	return nil
}

func (si *snmpInlet) Close() error {
	return nil
}

func (si *snmpInlet) Interval() time.Duration {
	return si.interval
}

func (si *snmpInlet) Process(next engine.InletNextFunc) {
	if si.runCountLimit > 0 && si.runCount > si.runCountLimit {
		next(nil, io.EOF)
		return
	}
	recs, err := si.Gather()
	si.runCount++
	if err == nil && si.runCountLimit > 0 && si.runCount >= si.runCountLimit {
		err = io.EOF
	}
	next(recs, err)
}

func (si *snmpInlet) Gather() ([]engine.Record, error) {
	var wg sync.WaitGroup
	var result []engine.Record
	var resultLock sync.Mutex

	for i, agent := range si.agents {
		wg.Add(1)
		go func(idx int, agent string) {
			defer wg.Done()
			gs, err := si.getConnection(idx)
			if err != nil {
				si.ctx.LogWarn("inlets.snmp", "connecting", agent, "error", err)
				return
			}
			t := Table{
				Name:   si.name,
				Fields: si.fields,
			}
			if recs, err := si.gatherTable(gs, t, false); err != nil {
				si.ctx.LogWarn("inlets.snmp", "gathering table", si.name, "error", err)
			} else {
				resultLock.Lock()
				result = append(result, recs...)
				resultLock.Unlock()
			}

			for _, table := range si.tables {
				if recs, err := si.gatherTable(gs, table, true); err != nil {
					si.ctx.LogWarn("inlets.snmp", "gathering table", table.Name, "error", err)
				} else {
					resultLock.Lock()
					result = append(result, recs...)
					resultLock.Unlock()
				}
			}
		}(i, agent)
	}
	wg.Wait()
	return result, nil
}

func (si *snmpInlet) gatherTable(gs Conn, table Table, walk bool) ([]engine.Record, error) {
	rt, err := table.Build(gs, walk)
	if err != nil {
		return nil, err
	}

	ret := make([]engine.Record, 0, len(rt.Rows))
	for _, tr := range rt.Rows {
		rec := engine.NewRecord()

		if !walk {
			for k, v := range tr.Tags {
				rec.Tags().Set(k, engine.NewValue(v))
			}
		} else {
			for _, k := range table.InheritTags {
				if v, ok := tr.Tags[k]; ok {
					rec.Tags().Set(k, engine.NewValue(v))
				}
			}
		}

		for name, value := range tr.Fields {
			recName := table.Name
			if idx, ok := tr.Tags["ifIndex"]; ok {
				recName += si.nameInfix + idx
			} else {
				si.ctx.LogDebug("snmp ifIndex not found", "tags", tr.Tags, "fields", tr.Fields)
			}
			recName = recName + si.nameInfix + name
			switch v := value.(type) {
			case string:
				rec.Append(engine.NewField(recName, v))
			case float64:
				rec.Append(engine.NewField(recName, v))
			case int64:
				rec.Append(engine.NewField(recName, v))
			case bool:
				rec.Append(engine.NewField(recName, v))
			case float32:
				rec.Append(engine.NewField(recName, float64(v)))
			case int:
				rec.Append(engine.NewField(recName, int64(v)))
			case uint32:
				rec.Append(engine.NewField(recName, uint64(v)))
			case uint:
				rec.Append(engine.NewField(recName, uint64(v)))
			default:
				si.ctx.LogWarn("inlet_snmp drop record", "name", recName, "type", fmt.Sprintf("%T", v), "value", fmt.Sprintf("%v", v))
			}
		}
		if len(rec.Names()) > 0 {
			ret = append(ret, rec)
		}
	}
	return ret, nil
}

func (si *snmpInlet) getConnection(idx int) (Conn, error) {
	if gs := si.connectionCache[idx]; gs != nil {
		if err := gs.Reconnect(); err != nil {
			return gs, fmt.Errorf("reconnecting: %w", err)
		}
		return gs, nil
	}
	agent := si.agents[idx]
	gs, err := NewWrapper(si.clientConf)
	if err != nil {
		return nil, err
	}
	err = gs.SetAgent(agent)
	if err != nil {
		return nil, err
	}
	si.connectionCache[idx] = gs

	if err := gs.Connect(); err != nil {
		return gs, fmt.Errorf("set up connecting: %w", err)
	}

	return gs, nil
}

type ClientConfig struct {
	Timeout                 time.Duration
	Retries                 int
	Version                 int
	UseUnconnectedUDPSocket bool

	// Version 1 and 2
	Community string

	// Version 2 and 3
	MaxRepetitions uint32

	// Version 3 only
	ContextName  string
	SecLevel     string
	SecName      string
	AuthProtocol string
	AuthPassword string
	PrivProtocol string
	PrivPassword string
	EngineID     string
	EngineBoots  uint32
	EngineTime   uint32
}
