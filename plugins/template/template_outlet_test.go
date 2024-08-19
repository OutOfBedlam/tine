package template_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	_ "github.com/OutOfBedlam/tine/plugins/base"
	_ "github.com/OutOfBedlam/tine/plugins/template"
	"github.com/stretchr/testify/require"
)

func TestTemplate(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{
			input:  "./testdata/temp1.toml",
			expect: "./testdata/temp1.txt",
		},
		{
			input:  "./testdata/temp2.toml",
			expect: "./testdata/temp2.txt",
		},
		{
			input:  "./testdata/temp3.toml",
			expect: "./testdata/temp3.txt",
		},
	}

	for _, tt := range tests {
		input, _ := os.ReadFile(tt.input)
		out := &strings.Builder{}
		seq := int64(0)
		engine.Now = func() time.Time { seq++; return time.Unix(1724045000+seq, 0) }
		p, err := engine.New(engine.WithConfig(string(input)), engine.WithWriter(out))
		if err != nil {
			t.Fatal(err)
		}
		if err := p.Run(); err != nil {
			t.Log("Fail:", tt.input)
			t.Fatal(err)
		}
		expect, _ := os.ReadFile(tt.expect)
		expect = bytes.ReplaceAll(expect, []byte{'\r'}, nil)
		result := strings.ReplaceAll(out.String(), "\r", "")
		require.Equal(t, string(expect), result, "input=%s", tt.input)
	}
}

func TestTemplate_file(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{
			input:  "./testdata/temp4.toml",
			expect: "./testdata/temp4.txt",
		},
	}

	for _, tt := range tests {
		input, _ := os.ReadFile(tt.input)
		seq := int64(0)
		engine.Now = func() time.Time { seq++; return time.Unix(1724045000+seq, 0) }
		p, err := engine.New(engine.WithConfig(string(input)))
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove("./testdata/_out.txt")
		err = p.Run()
		if err != nil {
			t.Log("Fail:", tt.input)
			t.Fatal(err)
		}
		result, _ := os.ReadFile("./testdata/_out.txt")
		expect, _ := os.ReadFile(tt.expect)
		require.EqualValues(t, expect, result, "input=%s", tt.input)
	}
}
