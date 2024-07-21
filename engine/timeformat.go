package engine

import (
	"strconv"
	"time"
)

type Timeformatter struct {
	format string
	loc    *time.Location
}

func NewTimeformatter(format string) *Timeformatter {
	return &Timeformatter{format: format, loc: time.Local}
}

func NewTimeformatterWithLocation(format string, tz *time.Location) *Timeformatter {
	return &Timeformatter{format: format, loc: tz}
}

var DefaultTimeformatter = &Timeformatter{format: time.RFC3339, loc: time.Local}

func (tf *Timeformatter) IsEpoch() bool {
	switch tf.format {
	case "ns", "us", "ms", "s":
		return true
	default:
		return false
	}
}

func (tf *Timeformatter) Epoch(t time.Time) int64 {
	switch tf.format {
	case "ns":
		return t.UnixNano()
	case "us":
		return t.UnixNano() / 1e3
	case "ms":
		return t.UnixNano() / 1e6
	case "s":
		return t.Unix()
	default:
		return t.Unix()
	}
}

func (tf *Timeformatter) Format(t time.Time) string {
	switch tf.format {
	case "ns":
		return strconv.FormatInt(t.UnixNano(), 10)
	case "us":
		return strconv.FormatInt(t.UnixNano()/1e3, 10)
	case "ms":
		return strconv.FormatInt(t.UnixNano()/1e6, 10)
	case "s":
		return strconv.FormatInt(t.Unix(), 10)
	default:
		return t.In(tf.loc).Format(tf.format)
	}
}

func (tf *Timeformatter) Parse(str string) (time.Time, error) {
	switch tf.format {
	case "ns":
		ns, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(0, ns), nil
	case "us":
		us, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(0, us*1e3), nil
	case "ms":
		ms, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(0, ms*1e6), nil
	case "s":
		s, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(s, 0), nil
	default:
		return time.ParseInLocation(tf.format, str, tf.loc)
	}
}
