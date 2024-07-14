package snmp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/models"
	"github.com/sleepinggenius2/gosmi/types"
)

type gosmiTranslator struct {
}

func NewGosmiTranslator(paths []string) (*gosmiTranslator, error) {
	err := LoadMibsFromPath(paths, &GosmiMibLoader{})
	if err == nil {
		return &gosmiTranslator{}, nil
	}
	return nil, err
}

func (t *gosmiTranslator) SnmpTranslate(oid string) (
	mibName string, oidNum string, oidText string,
	conversion string,
	err error,
) {
	mibName, oidNum, oidText, conversion, _, err = snmpTranslateCall(oid)
	return mibName, oidNum, oidText, conversion, err
}

func (t *gosmiTranslator) SnmpTable(oid string) (
	mibName string, oidNum string, oidText string,
	fields []Field,
	err error) {
	mibName, oidNum, oidText, _, node, err := snmpTranslateCall(oid)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("translating: %w", err)
	}

	mibPrefix := mibName + "::"

	col, tagOids := getIndex(mibPrefix, node)
	for _, c := range col {
		_, isTag := tagOids[mibPrefix+c]
		fields = append(fields, Field{Name: c, Oid: mibPrefix + c, IsTag: isTag})
	}

	return mibName, oidNum, oidText, fields, nil
}

func (t *gosmiTranslator) SnmpFormatEnum(oid string, value any, full bool) (
	formatted string,
	err error,
) {
	_, _, _, _, node, err := snmpTranslateCall(oid)

	if err != nil {
		return "", err
	}

	var v models.Value
	if full {
		v = node.FormatValue(value, models.FormatEnumName, models.FormatEnumValue)
	} else {
		v = node.FormatValue(value, models.FormatEnumName)
	}

	return v.Formatted, nil
}

func getIndex(mibPrefix string, node gosmi.SmiNode) (col []string, tagOids map[string]struct{}) {
	tagOids = make(map[string]struct{})

	for _, index := range node.GetIndex() {
		tagOids[mibPrefix+index.Name] = struct{}{}
	}

	_, col = node.GetColumns()
	return col, tagOids
}

func snmpTranslateCall(oid string) (
	mibName string, oidNum string, oidText string, conversion string, node gosmi.SmiNode, err error,
) {
	var out gosmi.SmiNode
	var end string
	if strings.ContainsAny(oid, "::") {
		s := strings.SplitN(oid, "::", 2)
		moduleName := s[0]
		module, err := gosmi.GetModule(moduleName)
		if err != nil {
			return oid, oid, oid, "", gosmi.SmiNode{}, err
		}
		if s[1] == "" {
			return "", oid, oid, "", gosmi.SmiNode{}, fmt.Errorf("cannot parse %v", oid)
		}
		// node becomes sysUpTime.0
		node := s[1]
		if strings.ContainsAny(node, ".") {
			s := strings.SplitN(node, ".", 2)
			node = s[0]
			end = s[1]
		}

		out, err = module.GetNode(node)
		if err != nil {
			return oid, oid, oid, "", out, err
		}
		if oidNum = out.RenderNumeric(); oidNum == "" {
			return oid, oid, oid, "", out, fmt.Errorf("cannot render numeric %v", oid)
		}
		oidNum = "." + oidNum + end
	} else if strings.ContainsAny(oid, "abcdefghijklnmopqrstuvwxyz") {
		// handle mixed oid ex. .iso.2.3
		s := strings.Split(oid, ".")
		for i := range s {
			if strings.ContainsAny(s[i], "abcdefghijklnmopqrstuvwxyz") {
				out, err = gosmi.GetNode(s[i])
				if err != nil {
					return oid, oid, oid, "", out, err
				}
				s[i] = out.RenderNumeric()
			}
		}
		oidNum = strings.Join(s, ".")
		out, err = gosmi.GetNodeByOID(types.OidMustFromString(oidNum))
		if err != nil {
			return oid, oid, oid, "", out, err
		}
	} else {
		out, err = gosmi.GetNodeByOID(types.OidMustFromString(oid))
		oidNum = oid
		if err != nil || out.Name == "iso" {
			return oid, oid, oid, "", out, nil
		}
	}
	tc := out.GetSubtree()

	for i := range tc {
		if tc[i].Type == nil {
			break
		}
		switch tc[i].Type.Name {
		case "MacAddress", "PhysAddress":
			conversion = "hwaddr"
		case "InetAddressIPv4", "InetAddressIPv6", "InetAddress", "IPSIpAddress":
			conversion = "ipaddr"
		}
	}

	oidText = out.RenderQualified()
	i := strings.Index(oidText, "::")
	if i == -1 {
		return "", oid, oid, "", out, errors.New("not found")
	}
	mibName = oidText[:i]
	oidText = oidText[i+2:] + end

	return mibName, oidNum, oidText, conversion, out, nil
}

type MibEntry struct {
	MibName string
	OidText string
}
