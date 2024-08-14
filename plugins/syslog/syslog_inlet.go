package syslog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"unicode"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/leodido/go-syslog/v4"
	"github.com/leodido/go-syslog/v4/nontransparent"
	"github.com/leodido/go-syslog/v4/octetcounting"
	"github.com/leodido/go-syslog/v4/rfc3164"
	"github.com/leodido/go-syslog/v4/rfc5424"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "syslog",
		Factory: SyslogInlet,
	})
}

func SyslogInlet(ctx *engine.Context) engine.Inlet {
	return &syslogInlet{
		ctx:    ctx,
		pushCh: make(chan Data),
	}
}

type syslogInlet struct {
	ctx *engine.Context

	separator string

	pushCh    chan Data
	closeOnce sync.Once
	closed    bool
	closeWg   sync.WaitGroup

	// tcp
	lsnr      net.Listener
	ctxCancel context.CancelFunc
	// udp
	pktConn net.PacketConn
}

var _ = engine.Inlet((*syslogInlet)(nil))

type Data struct {
	reccords []engine.Record
	err      error
}

func (si *syslogInlet) Apply() int {
	return 2024
}

func (si *syslogInlet) Open() error {
	address := si.ctx.Config().GetString("address", "tcp://127.0.0.1:6514")
	protoAddr := strings.SplitN(address, "://", 2)
	si.separator = si.ctx.Config().GetString("separator", "_")

	slog.Debug("inlet-syslog", "address", address)

	switch protoAddr[0] {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		if ln, err := net.Listen(protoAddr[0], protoAddr[1]); err != nil {
			return err
		} else {
			si.lsnr = ln
		}
		si.closeWg.Add(1)
		go si.handleStream()
	case "udp", "udp4", "udp6", "ip", "ip4", "ip6", "unixgram":
		lc := &net.ListenConfig{}

		ctx, cancel := context.WithCancel(si.ctx)
		if ln, err := lc.ListenPacket(ctx, protoAddr[0], protoAddr[1]); err != nil {
			cancel()
			return err
		} else {
			si.ctxCancel = cancel
			si.pktConn = ln
		}
		si.closeWg.Add(1)
		go si.handleDiagram()

	default:
		return fmt.Errorf("unsupported protocol: %s in %s", protoAddr[0], address)
	}

	return nil
}

func (si *syslogInlet) Close() error {
	si.closed = true
	si.closeOnce.Do(func() {
		si.ctxCancel()
		if si.lsnr != nil {
			si.lsnr.Close()
		}
		if si.pktConn != nil {
			si.pktConn.Close()
		}
		if si.pushCh != nil {
			close(si.pushCh)
		}
		si.closeWg.Wait()
	})
	return nil
}

func (si *syslogInlet) Process(next engine.InletNextFunc) {
	for d := range si.pushCh {
		next(d.reccords, d.err)
	}
}

func (si *syslogInlet) handleDiagram() {
	defer si.closeWg.Done()
	parallelism := si.ctx.Config().GetInt("parallelism", 1)
	syslogStandard := si.ctx.Config().GetString("syslog_standard", "rfc3164")

	var parser syslog.Machine
	switch strings.ToUpper(syslogStandard) {
	case "RFC3164":
		parser = rfc3164.NewParser(rfc3164.WithYear(rfc3164.CurrentYear{}))
	case "RFC5452":
		parser = rfc5424.NewParser()
	}
	bestEffort := si.ctx.Config().GetBool("best_effort", false)
	if bestEffort {
		parser.WithBestEffort()
	}

	sem := make(chan struct{}, parallelism)
	for !si.closed {
		sem <- struct{}{}
		si.closeWg.Add(1)
		go func() {
			defer func() {
				<-sem
				si.closeWg.Done()
			}()
			buf := make([]byte, 64*1024)
			n, addr, err := si.pktConn.ReadFrom(buf)
			if err != nil {
				if !si.closed {
					slog.Warn("inlet-syslog", "read error", err)
					si.pushCh <- Data{err: err}
				}
				return
			}
			message, err := parser.Parse(buf[:n])
			if err != nil {
				slog.Warn("inlet-syslog", "parse error", err)
				return
			}
			if message == nil {
				return
			}
			if r := records(message, si.separator); r != nil {
				r = r.Append(engine.NewField("remote_host", addr.(*net.UDPAddr).IP.String()))
				si.pushCh <- Data{reccords: []engine.Record{r}}
			}
		}()
	}
}

func (si *syslogInlet) handleStream() {
	defer si.closeWg.Done()
	var parser syslog.Parser
	bestEffort := si.ctx.Config().GetBool("best_effort", false)
	framing := si.ctx.Config().GetString("framing", "octetcounting")
	opts := []syslog.ParserOption{}
	if bestEffort {
		opts = append(opts, syslog.WithBestEffort())
	}
	switch framing {
	case "octetcounting":
		parser = octetcounting.NewParser(opts...)
	case "non-transport":
		parser = nontransparent.NewParser(opts...)
	}
	parser.WithListener(func(r *syslog.Result) {
		if r.Error != nil {
			slog.Warn("inlet-syslog", "parse error", r.Error)
			return
		}
		if r.Message == nil {
			return
		}
		if r := records(r.Message, si.separator); r != nil {
			si.pushCh <- Data{reccords: []engine.Record{r}}
		}
	})
	si.pushCh <- Data{err: io.EOF}
}

func records(msg syslog.Message, separator string) engine.Record {
	ret := engine.NewRecord()
	switch msg := msg.(type) {
	default:
		slog.Warn("inlet-syslog", "unsupported message type", fmt.Sprintf("%T", msg))
	case *rfc3164.SyslogMessage:
		ret = ret.Append(
			engine.NewField("facility_code", int64(*msg.Facility)),
			engine.NewField("severity_code", int64(*msg.Severity)),
		)
		if msg.Timestamp != nil {
			ret = ret.Append(engine.NewField("timestamp", *msg.Timestamp))
		}
		if msg.Hostname != nil {
			ret = ret.Append(engine.NewField("hostname", *msg.Hostname))
		}
		if msg.Appname != nil {
			ret = ret.Append(engine.NewField("appname", *msg.Appname))
		}
		if msg.ProcID != nil {
			ret = ret.Append(engine.NewField("procid", *msg.ProcID))
		}
		if msg.MsgID != nil {
			ret = ret.Append(engine.NewField("msgid", *msg.MsgID))
		}
		if msg.Message != nil {
			ret = ret.Append(engine.NewField(
				"message",
				strings.TrimRightFunc(*msg.Message, func(r rune) bool {
					return unicode.IsSpace(r)
				})))
		}
	case *rfc5424.SyslogMessage:
		// <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID [SD-ID STRUCTURED-DATA] MESSAGE
		ret = ret.Append(
			engine.NewField("facility_code", int64(*msg.Facility)),
			engine.NewField("severity_code", int64(*msg.Severity)),
			engine.NewField("version", int64(msg.Version)),
		)
		if msg.Timestamp != nil {
			ret = ret.Append(engine.NewField("timestamp", *msg.Timestamp))
		}
		if msg.Hostname != nil {
			ret = ret.Append(engine.NewField("hostname", *msg.Hostname))
		}
		if msg.Appname != nil {
			ret = ret.Append(engine.NewField("appname", *msg.Appname))
		}
		if msg.ProcID != nil {
			ret = ret.Append(engine.NewField("procid", *msg.ProcID))
		}
		if msg.MsgID != nil {
			ret = ret.Append(engine.NewField("msgid", *msg.MsgID))
		}
		if msg.Message != nil {
			ret = ret.Append(engine.NewField(
				"message",
				strings.TrimRightFunc(*msg.Message, func(r rune) bool {
					return unicode.IsSpace(r)
				})))
		}
		if msg.StructuredData != nil {
			for sdid, sdparams := range *msg.StructuredData {
				if len(sdparams) == 0 {
					// When SD-ID does not have params we indicate its presence with a bool
					ret = ret.Append(engine.NewField(sdid, true))
					continue
				}
				for k, v := range sdparams {
					ret = ret.Append(engine.NewField(sdid+separator+k, v))
				}
			}
		}
	}
	return ret
}
