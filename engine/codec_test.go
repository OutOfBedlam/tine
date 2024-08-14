package engine_test

import (
	"testing"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/OutOfBedlam/tine/plugins/base"
	"github.com/stretchr/testify/require"
)

func TestCodecs(t *testing.T) {
	engine.RegisterEncoder(&engine.EncoderReg{
		Name:        "test-json",
		Factory:     base.NewJSONEncoder,
		ContentType: "application/x-ndjson",
	})
	engine.RegisterEncoder(&engine.EncoderReg{
		Name:        "test-csv",
		Factory:     base.NewCSVEncoder,
		ContentType: "text/csv",
	})

	engine.RegisterDecoder(&engine.DecoderReg{
		Name:    "test-json",
		Factory: base.NewJSONDecoder,
	})
	engine.RegisterDecoder(&engine.DecoderReg{
		Name:    "test-csv",
		Factory: base.NewCSVDecoder,
	})

	enc := engine.GetEncoder("test-json")
	if enc == nil {
		t.Fatal("json encoder not found")
	}
	if enc.ContentType != "application/x-ndjson" {
		t.Fatalf("unexpected content type: %s", enc.ContentType)
	}
	dec := engine.GetDecoder("test-csv")
	if dec == nil {
		t.Fatal("csv encoder not found")
	}

	enc = engine.GetEncoder("test-csv")
	if enc == nil {
		t.Fatal("csv encoder not found")
	}
	if enc.ContentType != "text/csv" {
		t.Fatalf("unexpected content type: %s", enc.ContentType)
	}

	dec = engine.GetDecoder("test-json")
	if dec == nil {
		t.Fatal("json encoder not found")
	}

	names := engine.EncoderNames()
	require.Equal(t, []string{"csv", "json", "test-csv", "test-json"}, names)

	names = engine.DecoderNames()
	require.Equal(t, []string{"csv", "json", "test-csv", "test-json"}, names)

	engine.UnregisterEncoder("test-json")
	enc = engine.GetEncoder("test-json")
	if enc != nil {
		t.Fatal("json encoder should be unregistered")
	}

	names = engine.EncoderNames()
	require.Equal(t, []string{"csv", "json", "test-csv"}, names)

	engine.UnregisterDecoder("test-csv")
	dec = engine.GetDecoder("test-csv")
	if dec != nil {
		t.Fatal("csv encoder should be unregistered")
	}

	names = engine.DecoderNames()
	require.Equal(t, []string{"csv", "json", "test-json"}, names)

	engine.UnregisterEncoder("test-csv")
	enc = engine.GetEncoder("test-json")
	if enc != nil {
		t.Fatal("json encoder should be unregistered")
	}

	names = engine.EncoderNames()
	require.Equal(t, []string{"csv", "json"}, names)
}
