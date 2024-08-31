package fakeit

import (
	"io"
	"math/rand/v2"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/brianvoe/gofakeit/v7"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "fakeit",
		Factory: FakeItInlet,
	})
}

func FakeItInlet(ctx *engine.Context) engine.Inlet {
	return &fakeItInlet{ctx: ctx}
}

type fakeItInlet struct {
	ctx        *engine.Context
	faker      *gofakeit.Faker
	countLimit int
	count      int
	interval   time.Duration
	fields     []string
}

var _ = engine.Inlet((*fakeItInlet)(nil))

func (fi *fakeItInlet) Open() error {
	conf := fi.ctx.Config()
	fi.countLimit = conf.GetInt("count", 0)
	fi.interval = conf.GetDuration("interval", 10*time.Second)
	fi.fields = conf.GetStringSlice("fields", nil)
	seed := conf.GetInt64("seed", 0)
	gofakeit.Seed(seed)
	fi.faker = gofakeit.NewFaker(rand.NewPCG(11, 11), true)
	return nil
}

func (fi *fakeItInlet) Close() error {
	return nil
}

func (fi *fakeItInlet) Interval() time.Duration {
	return fi.interval
}

func (fi *fakeItInlet) Process(next engine.InletNextFunc) {
	if fi.countLimit > 0 && fi.count > fi.countLimit {
		next(nil, io.EOF)
		return
	}
	fields := make([]*engine.Field, 0, len(fi.fields))
	for _, field := range fi.fields {
		fields = append(fields, fi.gen(field))
	}
	rec := engine.NewRecord(fields...)

	fi.count++
	if fi.countLimit > 0 && fi.count >= fi.countLimit {
		next([]engine.Record{rec}, io.EOF)
	} else {
		next([]engine.Record{rec}, nil)
	}
}

func (fi *fakeItInlet) gen(n string) *engine.Field {
	switch n {
	case "name":
		return engine.NewField(n, fi.faker.Name())
	case "email":
		return engine.NewField(n, fi.faker.Email())
	case "phone":
		return engine.NewField(n, fi.faker.Phone())
	case "city":
		return engine.NewField(n, fi.faker.City())
	case "state":
		return engine.NewField(n, fi.faker.State())
	case "zip":
		return engine.NewField(n, fi.faker.Zip())
	case "country":
		return engine.NewField(n, fi.faker.Country())
	case "latitude":
		return engine.NewField(n, fi.faker.Latitude())
	case "longitude":
		return engine.NewField(n, fi.faker.Longitude())
	case "int":
		return engine.NewField(n, int64(fi.faker.Int64()))
	case "uint":
		return engine.NewField(n, uint64(fi.faker.Uint64()))
	case "float":
		return engine.NewField(n, fi.faker.Float64())
		// misc.
	case "uuid":
		return engine.NewField(n, fi.faker.UUID())
		// string
	case "digit":
		return engine.NewField(n, fi.faker.Digit())
	case "digitN":
		return engine.NewField(n, fi.faker.DigitN(5))
	case "letter":
		return engine.NewField(n, fi.faker.Letter())
	case "letterN":
		return engine.NewField(n, fi.faker.LetterN(5))
	case "lexify":
		return engine.NewField(n, fi.faker.Lexify("????"))
	case "numerify":
		return engine.NewField(n, fi.faker.Numerify("###"))
	default:
		return engine.NewField(n, n)
	}
}
