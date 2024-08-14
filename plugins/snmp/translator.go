package snmp

type Translator interface {
	SnmpTranslate(string) (
		mibName string, oidNum string, oidText string,
		conversion string,
		err error,
	)
	SnmpTable(oid string) (
		mibName string, oidNum string, oidText string,
		fields []Field,
		err error,
	)
	SnmpFormatEnum(oid string, value any, full bool) (
		formatted string,
		err error,
	)
}

var _ Translator = ((*gosmiTranslator)(nil))
var _ Translator = ((*netsnmpTranslator)(nil))
