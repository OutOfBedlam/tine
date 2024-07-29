package snmp

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "snmp",
		Factory: SnmpInlet,
		// Usage:   "--in-snmp <param>",
		// Help: []string{
		// 	"Gather metrics from SNMP agents.",
		// 	"Avaliable params:",
		// 	"    name=<string>      SNMP client name (default: in-snmp)",
		// 	"    agent=<string>     SNMP agent address, multiple agents are allowed",
		// 	"                       (e.g. agent=udp://127.0.0.1:161,agent=tcp://127.0.0.1:161)",
		// 	"    version=[1|2|3]    SNMP version (default: 2)",
		// 	"    community=<string> SNMP community string (default: public)",
		// 	"    unconnected_udp_socket=[true|false]  Use unconnected UDP socket (default: false)",
		// 	"    retries=<int>      SNMP retries (default: 3)",
		// 	"    max_repetitions=<int>  SNMP max repetitions (default: 10)",
		// 	"    context_name=<string>  SNMPv3 context name",
		// 	"    timeout=<duration>   SNMP timeout (default: 5s)",
		// 	"    sec_name=<string>    SNMPv3 security name",
		// 	"    sec_level=<lvl>      SNMPv3 security level (default: authNoPriv)",
		// 	"                         [noAuthNoPriv|authNoPriv|authPriv]",
		// 	"    auth_protocol=<alg>  SNMPv3 authentication protocol",
		// 	"                         [MD5|SHA|SHA224|SHA256|SHA384|SHA512]",
		// 	"    auth_password=<string>  SNMPv3 authentication password",
		// 	"    priv_protocol=<alg>  SNMPv3 privacy protocol",
		// 	"                         [DES|AES|AES192|AES192C|AES256|AES256C]",
		// 	"    priv_password=<string>  SNMPv3 privacy password",
		// 	"    translator=<string>  MIB translator (default: gosmi)",
		// 	"                         [gosmi|netsnmp]",
		// 	"    table=<string>     SNMP table name",
		// 	"    field=<string>     SNMP field name",
		// 	"    inherit_tags=<string>  Inherit tags from the table",
		// 	"    engine_id=<string>  SNMPv3 engine ID",
		// 	"    engine_boots=<int>  SNMPv3 engine boots",
		// 	"    engine_time=<int>   SNMPv3 engine time",
		// 	"    path=<string>       MIB path, multiple paths are allowed",
		// },
	})
}

func SnmpInlet(ctx *engine.Context) engine.Inlet {
	return &snmpInlet{ctx: ctx}
}

type snmpInlet struct {
	ctx             *engine.Context
	clientConf      ClientConfig
	name            string
	agents          []string
	tables          []Table
	fields          []Field
	connectionCache []Conn
	translator      Translator
}

var _ = engine.Inlet((*snmpInlet)(nil))

func (si *snmpInlet) Open() error {
	conf := si.ctx.Config()
	si.name = conf.GetString("name", "inlet.snmp")
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
	//si.clientConf.Path = conf.GetStringArray("path", []string{"/usr/share/snmp/mibs"})
	si.clientConf.Path = conf.GetStringSlice("path", []string{})
	si.clientConf.Translator = conf.GetString("translator", "gosmi")

	si.agents = conf.GetStringSlice("agent", nil)
	if len(si.agents) == 0 {
		return fmt.Errorf("no SNMP agent specified")
	}

	switch si.clientConf.Translator {
	case "gosmi":
		if trans, err := NewGosmiTranslator(si.clientConf.Path); err != nil {
			return err
		} else {
			si.translator = trans
		}
	case "netsnmp":
		si.translator = NewNetsnmpTranslator()
	default:
		return fmt.Errorf("invalid translator %q", si.clientConf.Translator)
	}

	si.connectionCache = make([]Conn, len(si.agents))

	for i := range si.tables {
		if err := si.tables[i].init(si.translator); err != nil {
			return fmt.Errorf("initializing table %s: %w", si.tables[i].Name, err)
		}
	}

	for i := range si.fields {
		if err := si.fields[i].Init(si.translator); err != nil {
			return fmt.Errorf("initializing field %s: %w", si.fields[i].Name, err)
		}
	}
	return nil
}

func (si *snmpInlet) Close() error {
	return nil
}

func (si *snmpInlet) Interval() time.Duration {
	return si.ctx.Config().GetDuration("interval", si.clientConf.Timeout)
}

func (si *snmpInlet) Process(next engine.InletNextFunc) {
	recs, err := si.Gather()
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
				slog.Warn("inlet_snmp", "connecting", agent, "error", err)
				return
			}
			t := Table{
				Name:   si.name,
				Fields: si.fields,
			}
			topTags := map[string]string{}
			if recs, err := si.gatherTable(gs, t, topTags, false); err != nil {
				slog.Warn("inlet_snmp", "gathering table", si.name, "error", err)
			} else {
				resultLock.Lock()
				result = append(result, recs...)
				resultLock.Unlock()
			}

			for _, table := range si.tables {
				if recs, err := si.gatherTable(gs, table, topTags, true); err != nil {
					slog.Warn("inlet_snmp", "gathering table", table.Name, "error", err)
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

func (si *snmpInlet) gatherTable(gs Conn, table Table, topTags map[string]string, walk bool) ([]engine.Record, error) {
	rt, err := table.Build(gs, walk)
	if err != nil {
		return nil, err
	}

	ret := make([]engine.Record, 0, len(rt.Rows))
	for _, tr := range rt.Rows {
		if !walk {
			for k, v := range tr.Tags {
				topTags[k] = v
			}
		} else {
			for _, k := range table.InheritTags {
				if v, ok := tr.Tags[k]; ok {
					topTags[k] = v
				}
			}
		}

		rec := engine.NewRecord()
		for name, value := range tr.Fields {
			recName := table.Name
			if idx, ok := tr.Tags["ifIndex"]; ok {
				recName += "." + idx
			} else {
				si.ctx.LogDebug("snmp ifIndex not found", "tags", tr.Tags, "fields", tr.Fields)
			}
			recName = recName + "." + name
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
				slog.Warn("inlet_snmp drop record", "name", recName, "type", fmt.Sprintf("%T", v), "value", fmt.Sprintf("%v", v))
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
