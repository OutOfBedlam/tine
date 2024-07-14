package snmp

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
)

type Conn interface {
	Host() string
	Walk(string, gosnmp.WalkFunc) error
	Get(oids []string) (*gosnmp.SnmpPacket, error)
	Reconnect() error
}

var _ = Conn((*GosnmpWrapper)(nil))

type GosnmpWrapper struct {
	*gosnmp.GoSNMP
}

func (gs GosnmpWrapper) Host() string {
	return gs.Target
}

func (gs GosnmpWrapper) Reconnect() error {
	if gs.Conn == nil {
		return gs.Connect()
	}
	return nil
}

func (gs GosnmpWrapper) Walk(oid string, f gosnmp.WalkFunc) error {
	if gs.Version == gosnmp.Version1 {
		return gs.GoSNMP.Walk(oid, f)
	}
	return gs.GoSNMP.BulkWalk(oid, f)
}

func (gs GosnmpWrapper) Get(oids []string) (*gosnmp.SnmpPacket, error) {
	return gs.GoSNMP.Get(oids)
}

// SetAgent sets the target agent for the connection.
// sheme://host:port
func (gs GosnmpWrapper) SetAgent(agent string) error {
	if !strings.Contains(agent, "://") {
		agent = "udp://" + agent
	}

	addr, err := url.Parse(agent)
	if err != nil {
		return err
	}
	// Only allow udp{4,6} and tcp{4,6}.
	// Allowing ip{4,6} does not make sense as specifying a port
	// requires the specification of a protocol.
	// gosnmp does not handle these errors well, which is why
	// they can result in cryptic errors by net.Dial.
	switch addr.Scheme {
	case "udp", "tcp", "udp4", "tcp4", "udp6", "tcp6":
		gs.Transport = addr.Scheme
	default:
		return fmt.Errorf("unsupported scheme %q", addr.Scheme)
	}
	gs.Target = addr.Hostname()
	if strPort := addr.Port(); strPort == "" {
		gs.Port = 161
	} else {
		if port, err := strconv.ParseUint(strPort, 10, 16); err != nil {
			return err
		} else {
			gs.Port = uint16(port)
		}
	}
	return nil
}

func NewWrapper(c ClientConfig) (GosnmpWrapper, error) {
	gs := GosnmpWrapper{&gosnmp.GoSNMP{}}
	gs.Timeout = c.Timeout
	gs.Retries = c.Retries
	gs.UseUnconnectedUDPSocket = c.UseUnconnectedUDPSocket

	switch c.Version {
	case 3:
		gs.Version = gosnmp.Version3
	case 2:
		gs.Version = gosnmp.Version2c
	case 1:
		gs.Version = gosnmp.Version1
	default:
		return gs, fmt.Errorf("unsupported SNMP version %d", c.Version)
	}

	if c.Version < 3 {
		if c.Community == "" {
			gs.Community = "public"
		} else {
			gs.Community = c.Community
		}
	}
	gs.MaxRepetitions = c.MaxRepetitions

	if c.Version != 3 {
		return gs, nil
	}

	// Version 3 specific settings
	gs.ContextName = c.ContextName
	sp := &gosnmp.UsmSecurityParameters{}
	gs.SecurityParameters = sp
	gs.SecurityModel = gosnmp.UserSecurityModel

	switch strings.ToLower(c.SecLevel) {
	case "noauthnopriv", "":
		gs.MsgFlags = gosnmp.NoAuthNoPriv
	case "authnopriv":
		gs.MsgFlags = gosnmp.AuthNoPriv
	case "authpriv":
		gs.MsgFlags = gosnmp.AuthPriv
	default:
		return gs, fmt.Errorf("unsupported security level %q", c.SecLevel)
	}
	sp.UserName = c.SecName

	switch strings.ToLower(c.AuthProtocol) {
	case "md5":
		sp.AuthenticationProtocol = gosnmp.MD5
	case "sha":
		sp.AuthenticationProtocol = gosnmp.SHA
	case "sha224":
		sp.AuthenticationProtocol = gosnmp.SHA224
	case "sha256":
		sp.AuthenticationProtocol = gosnmp.SHA256
	case "sha384":
		sp.AuthenticationProtocol = gosnmp.SHA384
	case "sha512":
		sp.AuthenticationProtocol = gosnmp.SHA512
	case "":
		sp.AuthenticationProtocol = gosnmp.NoAuth
	default:
		return gs, fmt.Errorf("unsupported authentication protocol %q", c.AuthProtocol)
	}

	sp.AuthenticationPassphrase = c.AuthPassword
	switch strings.ToLower(c.PrivProtocol) {
	case "des":
		sp.PrivacyProtocol = gosnmp.DES
	case "aes":
		sp.PrivacyProtocol = gosnmp.AES
	case "aes192":
		sp.PrivacyProtocol = gosnmp.AES192
	case "aes192c":
		sp.PrivacyProtocol = gosnmp.AES192C
	case "aes256":
		sp.PrivacyProtocol = gosnmp.AES256
	case "aes256c":
		sp.PrivacyProtocol = gosnmp.AES256C
	case "":
		sp.PrivacyProtocol = gosnmp.NoPriv
	default:
		return gs, fmt.Errorf("unsupported privacy protocol %q", c.PrivProtocol)
	}
	sp.PrivacyPassphrase = c.PrivPassword

	sp.AuthoritativeEngineID = c.EngineID
	sp.AuthoritativeEngineBoots = c.EngineBoots
	sp.AuthoritativeEngineTime = c.EngineTime

	return gs, nil
}
