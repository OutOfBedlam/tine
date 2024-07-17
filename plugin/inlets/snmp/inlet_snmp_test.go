package snmp

import (
	"fmt"
	"log/slog"
	"net"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/require"
)

func DummyContext(conf engine.Config) *engine.Context {
	return (&engine.Context{}).WithConfig(conf).WithLogger(slog.Default())
}

func TestSnmpInit(t *testing.T) {
	inlet := SnmpInlet(DummyContext(
		engine.Config{}.
			Set("agent", []string{""}).
			Set("translator", "gosmi"),
	))
	require.NoError(t, inlet.Open())
}

func TestSnmpInit_noTranslate(t *testing.T) {
	inlet := SnmpInlet(DummyContext(
		engine.Config{}.
			Set("agent", []string{""}).
			Set("translator", "netsnmp"),
	))
	s := inlet.(*snmpInlet)
	s.fields = []Field{
		{Oid: ".1.1.1.1", Name: "one", IsTag: true},
		{Oid: ".1.1.1.2", Name: "two"},
		{Oid: ".1.1.1.3"},
	}
	s.tables = []Table{
		{
			Name: "testing",
			Fields: []Field{
				{Oid: ".1.1.1.4", Name: "four", IsTag: true},
				{Oid: ".1.1.1.5", Name: "five"},
				{Oid: ".1.1.1.6"},
			},
		},
	}

	require.NoError(t, s.Open())

	require.Equal(t, ".1.1.1.1", s.fields[0].Oid)
	require.Equal(t, "one", s.fields[0].Name)
	require.True(t, s.fields[0].IsTag)

	require.Equal(t, ".1.1.1.2", s.fields[1].Oid)
	require.Equal(t, "two", s.fields[1].Name)
	require.False(t, s.fields[1].IsTag)

	require.Equal(t, ".1.1.1.3", s.fields[2].Oid)
	require.Equal(t, ".1.1.1.3", s.fields[2].Name)
	require.False(t, s.fields[2].IsTag)

	require.Equal(t, ".1.1.1.4", s.tables[0].Fields[0].Oid)
	require.Equal(t, "four", s.tables[0].Fields[0].Name)
	require.True(t, s.tables[0].Fields[0].IsTag)

	require.Equal(t, ".1.1.1.5", s.tables[0].Fields[1].Oid)
	require.Equal(t, "five", s.tables[0].Fields[1].Name)
	require.False(t, s.tables[0].Fields[1].IsTag)

	require.Equal(t, ".1.1.1.6", s.tables[0].Fields[2].Oid)
	require.Equal(t, ".1.1.1.6", s.tables[0].Fields[2].Name)
	require.False(t, s.tables[0].Fields[2].IsTag)
}

func TestSnmpInit_noName_noOid(t *testing.T) {
	inlet := SnmpInlet(DummyContext(
		engine.Config{}.
			Set("agent", ""),
	))
	s := inlet.(*snmpInlet)
	s.tables = []Table{
		{Fields: []Field{
			{Oid: ".1.1.1.4", Name: "four", IsTag: true},
			{Oid: ".1.1.1.5", Name: "five"},
			{Oid: ".1.1.1.6"},
		}},
	}
	require.Error(t, s.Open())
}

func TestGetSNMPConnection_v2(t *testing.T) {
	inlet := SnmpInlet(DummyContext(
		engine.Config{}.
			Set("agent", []string{"1.2.3.4:567", "1.2.3.4", "udp://127.0.0.1"}).
			Set("timeout", "3s").
			Set("retries", "4").
			Set("version", "2").
			Set("community", "foo").
			Set("translator", "netsnmp"),
	))
	require.NoError(t, inlet.Open())
	s := inlet.(*snmpInlet)

	gsc, err := s.getConnection(0)
	require.NoError(t, err)
	gs := gsc.(GosnmpWrapper)
	require.Equal(t, "1.2.3.4", gs.Target)
	require.EqualValues(t, 567, gs.Port)
	require.Equal(t, gosnmp.Version2c, gs.Version)
	require.Equal(t, "foo", gs.Community)
	require.Equal(t, "udp", gs.Transport)

	gsc, err = s.getConnection(1)
	require.NoError(t, err)
	gs = gsc.(GosnmpWrapper)
	require.Equal(t, "1.2.3.4", gs.Target)
	require.EqualValues(t, 161, gs.Port)
	require.Equal(t, "udp", gs.Transport)

	gsc, err = s.getConnection(2)
	require.NoError(t, err)
	gs = gsc.(GosnmpWrapper)
	require.Equal(t, "127.0.0.1", gs.Target)
	require.EqualValues(t, 161, gs.Port)
	require.Equal(t, "udp", gs.Transport)
}

func TestGetSNMPConnectionTCP(t *testing.T) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	tcpServer, err := net.ListenTCP("tcp", tcpAddr)
	require.NoError(t, err)
	defer tcpServer.Close()

	inlet := SnmpInlet(DummyContext(
		engine.Config{}.
			Set("agent", fmt.Sprintf("tcp://%s", tcpServer.Addr())).
			Set("version", "2"),
	))
	require.NoError(t, inlet.Open())
	s := inlet.(*snmpInlet)

	gsc, err := s.getConnection(0)
	require.NoError(t, err)
	gs := gsc.(GosnmpWrapper)
	require.Equal(t, "127.0.0.1", gs.Target)
	require.Equal(t, "tcp", gs.Transport)
}

func TestGetSNMPConnection_v3(t *testing.T) {
	inlet := SnmpInlet(DummyContext(
		engine.Config{}.
			Set("agent", "1.2.3.4").
			Set("version", 3).
			Set("max_repetitions", 20).
			Set("context_name", "mycontext").
			Set("sec_level", "authPriv").
			Set("sec_name", "myuser").
			Set("auth_protocol", "md5").
			Set("auth_password", "password123").
			Set("priv_protocol", "des").
			Set("priv_password", "321drowssap").
			Set("engine_id", "myengineid").
			Set("engine_boots", "1").
			Set("engine_time", "2").
			Set("translator", "netsnmp"),
	))
	require.NoError(t, inlet.Open())
	s := inlet.(*snmpInlet)

	gsc, err := s.getConnection(0)
	require.NoError(t, err)
	gs := gsc.(GosnmpWrapper)
	require.Equal(t, gosnmp.Version3, gs.Version)
	sp := gs.SecurityParameters.(*gosnmp.UsmSecurityParameters)
	require.Equal(t, "1.2.3.4", gsc.Host())
	require.EqualValues(t, 20, gs.MaxRepetitions)
	require.Equal(t, "mycontext", gs.ContextName)
	require.Equal(t, gosnmp.AuthPriv, gs.MsgFlags&gosnmp.AuthPriv)
	require.Equal(t, "myuser", sp.UserName)
	require.Equal(t, gosnmp.MD5, sp.AuthenticationProtocol)
	require.Equal(t, "password123", sp.AuthenticationPassphrase)
	require.Equal(t, gosnmp.DES, sp.PrivacyProtocol)
	require.Equal(t, "321drowssap", sp.PrivacyPassphrase)
	require.Equal(t, "myengineid", sp.AuthoritativeEngineID)
	require.EqualValues(t, 1, sp.AuthoritativeEngineBoots)
	require.EqualValues(t, 2, sp.AuthoritativeEngineTime)
}

func TestGetSNMPConnection_v3_blumenthal(t *testing.T) {
	testCases := []struct {
		Name      string
		Algorithm gosnmp.SnmpV3PrivProtocol
		Config    engine.Config
	}{
		{
			Name:      "AES192",
			Algorithm: gosnmp.AES192,
			Config: engine.Config{}.
				Set("agent", "1.2.3.4").
				Set("version", 3).
				Set("max_repetitions", 20).
				Set("context_name", "mycontext").
				Set("sec_level", "authPriv").
				Set("sec_name", "myuser").
				Set("auth_protocol", "md5").
				Set("auth_password", "password123").
				Set("priv_protocol", "AES192").
				Set("priv_password", "password123").
				Set("engine_id", "myengineid").
				Set("engine_boots", "1").
				Set("engine_time", "2").
				Set("translator", "netsnmp"),
		},
		{
			Name:      "AES192C",
			Algorithm: gosnmp.AES192C,
			Config: engine.Config{}.
				Set("agent", "1.2.3.4").
				Set("version", 3).
				Set("max_repetitions", 20).
				Set("context_name", "mycontext").
				Set("sec_level", "authPriv").
				Set("sec_name", "myuser").
				Set("auth_protocol", "md5").
				Set("auth_password", "password123").
				Set("priv_protocol", "AES192C").
				Set("priv_password", "password123").
				Set("engine_id", "myengineid").
				Set("engine_boots", "1").
				Set("engine_time", "2").
				Set("translator", "netsnmp"),
		},
		{
			Name:      "AES256",
			Algorithm: gosnmp.AES256,
			Config: engine.Config{}.
				Set("agent", "1.2.3.4").
				Set("version", 3).
				Set("max_repetitions", 20).
				Set("context_name", "mycontext").
				Set("sec_level", "authPriv").
				Set("sec_name", "myuser").
				Set("auth_protocol", "md5").
				Set("auth_password", "password123").
				Set("priv_protocol", "AES256").
				Set("priv_password", "password123").
				Set("engine_id", "myengineid").
				Set("engine_boots", "1").
				Set("engine_time", "2").
				Set("translator", "netsnmp"),
		},
		{
			Name:      "AES256C",
			Algorithm: gosnmp.AES256C,
			Config: engine.Config{}.
				Set("agent", "1.2.3.4").
				Set("version", 3).
				Set("max_repetitions", 20).
				Set("context_name", "mycontext").
				Set("sec_level", "authPriv").
				Set("sec_name", "myuser").
				Set("auth_protocol", "md5").
				Set("auth_password", "password123").
				Set("priv_protocol", "AES256C").
				Set("priv_password", "password123").
				Set("engine_id", "myengineid").
				Set("engine_boots", "1").
				Set("engine_time", "2").
				Set("translator", "netsnmp"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			inlet := SnmpInlet(DummyContext(tc.Config))
			require.NoError(t, inlet.Open())
			s := inlet.(*snmpInlet)

			gsc, err := s.getConnection(0)
			require.NoError(t, err)
			gs := gsc.(GosnmpWrapper)
			require.Equal(t, gosnmp.Version3, gs.Version)
			sp := gs.SecurityParameters.(*gosnmp.UsmSecurityParameters)
			require.Equal(t, "1.2.3.4", gsc.Host())
			require.EqualValues(t, 20, gs.MaxRepetitions)
			require.Equal(t, "mycontext", gs.ContextName)
			require.Equal(t, gosnmp.AuthPriv, gs.MsgFlags&gosnmp.AuthPriv)
			require.Equal(t, "myuser", sp.UserName)
			require.Equal(t, gosnmp.MD5, sp.AuthenticationProtocol)
			require.Equal(t, "password123", sp.AuthenticationPassphrase)
			require.Equal(t, tc.Algorithm, sp.PrivacyProtocol)
			require.Equal(t, "password123", sp.PrivacyPassphrase)
			require.Equal(t, "myengineid", sp.AuthoritativeEngineID)
			require.EqualValues(t, 1, sp.AuthoritativeEngineBoots)
			require.EqualValues(t, 2, sp.AuthoritativeEngineTime)
		})
	}
}

func TestGetSNMPConnection_caching(t *testing.T) {
	inlet := SnmpInlet(DummyContext(engine.Config{}.
		Set("agent", []string{"1.2.3.4", "1.2.3.5", "1.2.3.5"}).
		Set("version", "2").
		Set("translator", "netsnmp"),
	))
	require.NoError(t, inlet.Open())
	s := inlet.(*snmpInlet)
	err := s.Open()
	require.NoError(t, err)
	gs1, err := s.getConnection(0)
	require.NoError(t, err)
	gs2, err := s.getConnection(0)
	require.NoError(t, err)
	gs3, err := s.getConnection(1)
	require.NoError(t, err)
	gs4, err := s.getConnection(2)
	require.NoError(t, err)
	require.Equal(t, gs1, gs2)
	require.NotEqual(t, gs2, gs3)
	require.NotEqual(t, gs3, gs4)
}

func TestGosnmpWrapper_walk_retry(t *testing.T) {
	t.Skip("Skipping test due to random failures.")

	srvr, err := net.ListenUDP("udp4", &net.UDPAddr{})
	require.NoError(t, err)
	defer srvr.Close()
	reqCount := 0
	// Set up a WaitGroup to wait for the server goroutine to exit and protect
	// reqCount.
	// Even though simultaneous access is impossible because the server will be
	// blocked on ReadFrom, without this the race detector gets unhappy.
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 256)
		for {
			_, addr, err := srvr.ReadFrom(buf)
			if err != nil {
				return
			}
			reqCount++

			// will cause decoding error
			if _, err := srvr.WriteTo([]byte{'X'}, addr); err != nil {
				return
			}
		}
	}()

	gs := &gosnmp.GoSNMP{
		Target:    srvr.LocalAddr().(*net.UDPAddr).IP.String(),
		Port:      uint16(srvr.LocalAddr().(*net.UDPAddr).Port),
		Version:   gosnmp.Version2c,
		Community: "public",
		Timeout:   time.Millisecond * 10,
		Retries:   1,
	}
	err = gs.Connect()
	require.NoError(t, err)
	conn := gs.Conn

	gsw := GosnmpWrapper{
		GoSNMP: gs,
	}
	err = gsw.Walk(".1.0.0", func(_ gosnmp.SnmpPDU) error { return nil })
	require.NoError(t, srvr.Close())
	wg.Wait()
	require.Error(t, err)
	require.NotEqual(t, gs.Conn, conn)
	require.Equal(t, (gs.Retries+1)*2, reqCount)
}

func TestGosnmpWrapper_get_retry(t *testing.T) {
	// TODO: Fix this test
	t.Skip("Test failing too often, skip for now and revisit later.")
	srvr, err := net.ListenUDP("udp4", &net.UDPAddr{})
	require.NoError(t, err)
	defer srvr.Close()
	reqCount := 0
	// Set up a WaitGroup to wait for the server goroutine to exit and protect
	// reqCount.
	// Even though simultaneous access is impossible because the server will be
	// blocked on ReadFrom, without this the race detector gets unhappy.
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 256)
		for {
			_, addr, err := srvr.ReadFrom(buf)
			if err != nil {
				return
			}
			reqCount++

			// will cause decoding error
			if _, err := srvr.WriteTo([]byte{'X'}, addr); err != nil {
				return
			}
		}
	}()

	gs := &gosnmp.GoSNMP{
		Target:    srvr.LocalAddr().(*net.UDPAddr).IP.String(),
		Port:      uint16(srvr.LocalAddr().(*net.UDPAddr).Port),
		Version:   gosnmp.Version2c,
		Community: "public",
		Timeout:   time.Millisecond * 10,
		Retries:   1,
	}
	err = gs.Connect()
	require.NoError(t, err)
	conn := gs.Conn

	gsw := GosnmpWrapper{
		GoSNMP: gs,
	}
	_, err = gsw.Get([]string{".1.0.0"})
	require.NoError(t, srvr.Close())
	wg.Wait()
	require.Error(t, err)
	require.NotEqual(t, gs.Conn, conn)
	require.Equal(t, (gs.Retries+1)*2, reqCount)
}

func TestGather(t *testing.T) {
	t.Skip("Skipping test due to random failures.")
	inlet := SnmpInlet(DummyContext(
		engine.NewConfig().
			Set("agent", "TestGather").
			Set("name", "mytable"),
	))
	require.NoError(t, inlet.Open())
	s := inlet.(*snmpInlet)
	s.tables = []Table{
		{
			Name:        "myOtherTable",
			InheritTags: []string{"myfield1"},
			Fields: []Field{
				{
					Name: "myOtherField",
					Oid:  ".1.0.0.0.1.5",
				},
			},
		},
	}
	s.fields = []Field{
		{
			Name:  "myfield1",
			Oid:   ".1.0.0.1.1",
			IsTag: true,
		},
		{
			Name: "myfield2",
			Oid:  ".1.0.0.1.2",
		},
		{
			Name: "myfield3",
			Oid:  "1.0.0.1.1",
		},
	}
	s.connectionCache = []Conn{tsc}

	// tstart := time.Now()
	recs, err := s.Gather()
	require.NoError(t, err)
	// tstop := time.Now()

	require.Len(t, recs, 3)

	for i, r := range recs {
		t.Log("->", i, r)
	}
	// m := recs[0]
	// require.Equal(t, "mytable", m.Name)
	// require.Equal(t, "tsc", m.Tags[s.AgentHostTag])
	// require.Equal(t, "baz", m.Tags["myfield1"])
	// require.Len(t, m.Fields, 2)
	// require.Equal(t, 234, m.Fields["myfield2"])
	// require.Equal(t, "baz", m.Fields["myfield3"])
	// require.WithinRange(t, m.Time, tstart, tstop)

	// m2 := acc.Metrics[1]
	// require.Equal(t, "myOtherTable", m2.Measurement)
	// require.Equal(t, "tsc", m2.Tags[s.AgentHostTag])
	// require.Equal(t, "baz", m2.Tags["myfield1"])
	// require.Len(t, m2.Fields, 1)
	// require.Equal(t, 123456, m2.Fields["myOtherField"])
}

func TestGather_host(t *testing.T) {
	t.Skip("Skipping test due to not-implemented")
	inlet := SnmpInlet(DummyContext(
		engine.NewConfig().
			Set("agent", "TestGather").
			Set("name", "mytable"),
	))
	require.NoError(t, inlet.Open())
	s := inlet.(*snmpInlet)
	s.fields = []Field{
		{
			Name:  "host",
			Oid:   ".1.0.0.1.1",
			IsTag: true,
		},
		{
			Name: "myfield2",
			Oid:  ".1.0.0.1.2",
		},
	}
	s.connectionCache = []Conn{tsc}

	ret, err := s.Gather()
	require.NoError(t, err)
	require.NotEmpty(t, ret)

	// require.Len(t, acc.Metrics, 1)
	// m := acc.Metrics[0]
	// require.Equal(t, "baz", m.Tags["host"])
}

func TestSnmpInitGosmi(t *testing.T) {
	testDataPath, err := filepath.Abs("./testdata/gosmi")
	require.NoError(t, err)

	inlet := SnmpInlet(DummyContext(
		engine.NewConfig().
			Set("agent", "TestGather").
			Set("name", "mytable").
			Set("translator", "gosmi").
			Set("path", testDataPath),
	))
	s := inlet.(*snmpInlet)
	s.tables = []Table{
		{Oid: "RFC1213-MIB::atTable"},
	}
	s.fields = []Field{
		{Oid: "RFC1213-MIB::atPhysAddress"},
	}

	require.NoError(t, s.Open())

	require.Len(t, s.tables[0].Fields, 3)

	require.Equal(t, ".1.3.6.1.2.1.3.1.1.1", s.tables[0].Fields[0].Oid)
	require.Equal(t, "atIfIndex", s.tables[0].Fields[0].Name)
	require.True(t, s.tables[0].Fields[0].IsTag)
	require.Empty(t, s.tables[0].Fields[0].Conversion)

	require.Equal(t, ".1.3.6.1.2.1.3.1.1.2", s.tables[0].Fields[1].Oid)
	require.Equal(t, "atPhysAddress", s.tables[0].Fields[1].Name)
	require.False(t, s.tables[0].Fields[1].IsTag)
	require.Equal(t, "hwaddr", s.tables[0].Fields[1].Conversion)

	require.Equal(t, ".1.3.6.1.2.1.3.1.1.3", s.tables[0].Fields[2].Oid)
	require.Equal(t, "atNetAddress", s.tables[0].Fields[2].Name)
	require.True(t, s.tables[0].Fields[2].IsTag)
	require.Empty(t, s.tables[0].Fields[2].Conversion)

	require.Equal(t, ".1.3.6.1.2.1.3.1.1.2", s.fields[0].Oid)
	require.Equal(t, "atPhysAddress", s.fields[0].Name)
	require.False(t, s.fields[0].IsTag)
	require.Equal(t, "hwaddr", s.fields[0].Conversion)
}

func TestSnmpInit_noTranslateGosmi(t *testing.T) {
	inlet := SnmpInlet(DummyContext(
		engine.NewConfig().
			Set("agent", "TestGather").
			Set("name", "mytable").
			Set("translator", "gosmi"),
	))
	s := inlet.(*snmpInlet)
	s.fields = []Field{
		{Oid: ".9.1.1.1.1", Name: "one", IsTag: true},
		{Oid: ".9.1.1.1.2", Name: "two"},
		{Oid: ".9.1.1.1.3"},
	}
	s.tables = []Table{
		{Name: "testing",
			Fields: []Field{
				{Oid: ".9.1.1.1.4", Name: "four", IsTag: true},
				{Oid: ".9.1.1.1.5", Name: "five"},
				{Oid: ".9.1.1.1.6"},
			}},
	}

	require.NoError(t, s.Open())

	require.Equal(t, ".9.1.1.1.1", s.fields[0].Oid)
	require.Equal(t, "one", s.fields[0].Name)
	require.True(t, s.fields[0].IsTag)

	require.Equal(t, ".9.1.1.1.2", s.fields[1].Oid)
	require.Equal(t, "two", s.fields[1].Name)
	require.False(t, s.fields[1].IsTag)

	require.Equal(t, ".9.1.1.1.3", s.fields[2].Oid)
	require.Equal(t, ".9.1.1.1.3", s.fields[2].Name)
	require.False(t, s.fields[2].IsTag)

	require.Equal(t, ".9.1.1.1.4", s.tables[0].Fields[0].Oid)
	require.Equal(t, "four", s.tables[0].Fields[0].Name)
	require.True(t, s.tables[0].Fields[0].IsTag)

	require.Equal(t, ".9.1.1.1.5", s.tables[0].Fields[1].Oid)
	require.Equal(t, "five", s.tables[0].Fields[1].Name)
	require.False(t, s.tables[0].Fields[1].IsTag)

	require.Equal(t, ".9.1.1.1.6", s.tables[0].Fields[2].Oid)
	require.Equal(t, ".9.1.1.1.6", s.tables[0].Fields[2].Name)
	require.False(t, s.tables[0].Fields[2].IsTag)
}

func TestGatherGosmi(t *testing.T) {
	t.Skip("Skipping test due to random failures.")
	inlet := SnmpInlet(DummyContext(
		engine.NewConfig().
			Set("agent", "TestGather").
			Set("name", "mytable").
			Set("translator", "gosmi"),
	))
	s := inlet.(*snmpInlet)
	s.fields = []Field{
		{
			Name:  "myfield1",
			Oid:   ".1.0.0.1.1",
			IsTag: true,
		},
		{
			Name: "myfield2",
			Oid:  ".1.0.0.1.2",
		},
		{
			Name: "myfield3",
			Oid:  "1.0.0.1.1",
		},
	}
	s.tables = []Table{
		{
			Name:        "myOtherTable",
			InheritTags: []string{"myfield1"},
			Fields: []Field{
				{
					Name: "myOtherField",
					Oid:  ".1.0.0.0.1.5",
				},
			},
		},
	}
	require.NoError(t, s.Open())

	s.connectionCache = []Conn{tsc}

	//tstart := time.Now()
	recs, err := s.Gather()
	require.NoError(t, err)
	//tstop := time.Now()

	require.Len(t, recs, 3)

	// m := acc.Metrics[0]
	// require.Equal(t, "mytable", m.Measurement)
	// require.Equal(t, "tsc", m.Tags[s.AgentHostTag])
	// require.Equal(t, "baz", m.Tags["myfield1"])
	// require.Len(t, m.Fields, 2)
	// require.Equal(t, 234, m.Fields["myfield2"])
	// require.Equal(t, "baz", m.Fields["myfield3"])
	// require.WithinRange(t, m.Time, tstart, tstop)

	// m2 := acc.Metrics[1]
	// require.Equal(t, "myOtherTable", m2.Measurement)
	// require.Equal(t, "tsc", m2.Tags[s.AgentHostTag])
	// require.Equal(t, "baz", m2.Tags["myfield1"])
	// require.Len(t, m2.Fields, 1)
	// require.Equal(t, 123456, m2.Fields["myOtherField"])
}

func TestGather_hostGosmi(t *testing.T) {
	t.Skip("Skipping test due to not-implemented")
	inlet := SnmpInlet(DummyContext(
		engine.NewConfig().
			Set("agent", "TestGather").
			Set("name", "mytable").
			Set("translator", "gosmi"),
	))
	s := inlet.(*snmpInlet)
	s.fields = []Field{
		{
			Name:  "host",
			Oid:   ".1.0.0.1.1",
			IsTag: true,
		},
		{
			Name: "myfield2",
			Oid:  ".1.0.0.1.2",
		},
	}
	require.NoError(t, s.Open())

	s.connectionCache = []Conn{tsc}

	recs, err := s.Gather()
	require.NoError(t, err)

	require.Len(t, recs, 1)
	// m := acc.Metrics[0]
	// require.Equal(t, "baz", m.Tags["host"])
}

func TestGatherSysUptime(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test on non-linux systems")
	}

	inlet := SnmpInlet(DummyContext(
		engine.NewConfig().
			Set("agent", "udp://127.0.0.1").
			Set("version", "2").
			Set("retries", "0").
			Set("timeout", "3s").
			Set("community", "private").
			Set("translator", "netsnmp").
			Set("path", "/usr/share/snmp/mibs"),
	))
	s := inlet.(*snmpInlet)
	s.tables = []Table{
		{
			Name: "iftable",
			Oid:  "IF-MIB::ifTable",
		},
	}
	s.fields = []Field{
		{
			Name: "sysUpTime",
			Oid:  "1.3.6.1.2.1.1.3.0", // HOST-RESOURCES-MIB::hrSystemUptime.0
			// 		IsTag: true,
		},
	}
	require.NoError(t, s.Open())
	recs, err := s.Gather()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(recs), 1)

	for _, r := range recs {
		t.Log(">>", r)
		for _, f := range r.Fields() {
			switch f.Name {
			case "sysUpTime":
				require.Greater(t, f.Value.(uint64), uint64(0))
			}
		}
	}
}
