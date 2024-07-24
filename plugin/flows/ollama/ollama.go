package ollama

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterFlow(&engine.FlowReg{
		Name:    "ollama",
		Factory: OllamaFlow,
	})
}

func OllamaFlow(ctx *engine.Context) engine.Flow {
	return &ollamaFlow{ctx: ctx}
}

type ollamaFlow struct {
	ctx    *engine.Context
	addr   string
	client *http.Client
	model  string
	stream bool
}

func (of *ollamaFlow) Open() error {
	of.addr = of.ctx.Config().GetString("address", "")
	if !strings.HasPrefix(of.addr, "http://") && strings.HasPrefix(of.addr, "https://") {
		return fmt.Errorf("invalid address: %s", of.addr)
	}
	of.model = of.ctx.Config().GetString("model", "phi3")
	of.stream = of.ctx.Config().GetBool("stream", false)
	timeout := of.ctx.Config().GetDuration("timeout", 15*time.Second)

	of.ctx.LogDebug("flows.ollama", "address", of.addr, "timeout", timeout, "stream", of.stream)

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	of.client = &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return nil
}

func (of *ollamaFlow) Close() error {
	return nil
}

func (of *ollamaFlow) Parallelism() int {
	return 1
}

type GenerateRequest struct {
	Model   string           `json:"model"`
	Prompt  string           `json:"prompt"`
	Format  string           `json:"format,omitempty"`
	Stream  bool             `json:"stream"`
	Images  []string         `json:"images,omitempty"` // array of base64 encoded images
	Options *GenerateOptions `json:"options,omitempty"`
}

type GenerateOptions struct {
	NumKeep          int      `json:"num_keep"`          //: 5,
	Seed             int      `json:"seed"`              //: 42,
	NumPredit        int      `json:"num_predict"`       //: 100,
	TopK             int      `json:"top_k"`             //: 20,
	TopP             float32  `json:"top_p"`             //: 0.9,
	TfsZ             float32  `json:"tfs_z"`             //: 0.5,
	TypicalP         float32  `json:"typical_p"`         //: 0.7,
	RepeatLastN      int      `json:"repeat_last_n"`     //: 33,
	Temperature      float32  `json:"temperature"`       //: 0.8,
	RepeatPenalty    float32  `json:"repeat_penalty"`    //: 1.2,
	PresencePenalty  float32  `json:"presence_penalty"`  //: 1.5,
	FrequencyPenalty float32  `json:"frequency_penalty"` //: 1.0,
	Mirostat         float32  `json:"mirostat"`          //: 1,
	MirostatTau      float32  `json:"mirostat_tau"`      //: 0.8,
	MirostatEta      float32  `json:"mirostat_eta"`      //: 0.6,
	PenalizeNewline  bool     `json:"penalize_newline"`  //: true,
	Stop             []string `json:"stop"`              //: ["\n", "user:"],
	Numa             bool     `json:"numa"`              //: false,
	NumCtx           int      `json:"num_ctx"`           //: 1024,
	NumBatch         int      `json:"num_batch"`         //: 2,
	NumGpu           int      `json:"num_gpu"`           //: 1,
	MainGpu          int      `json:"main_gpu"`          //: 0,
	LowVram          bool     `json:"low_vram"`          //: false,
	F16KV            bool     `json:"f16_kv"`            //: true,
	VocabOnly        bool     `json:"vocab_only"`        //: false,
	UseMMAP          bool     `json:"use_mmap"`          //: true,
	UseMLock         bool     `json:"use_mlock"`         //: false,
	NumThread        int      `json:"num_thread"`        //: 8
}

type GenerateResponse struct {
	Error              string        `json:"error,omitempty"`
	Model              string        `json:"model"`
	Response           string        `json:"response"`
	Done               bool          `json:"done"`
	DoneReason         string        `json:"done_reason"`
	CreateAt           time.Time     `json:"created_at"`                     // "2024-07-22T11:34:39.896275469Z"
	TotalDuration      time.Duration `json:"total_duration,omitempty"`       //16537307773
	LoadDuration       time.Duration `json:"load_duration,omitempty"`        //14068211
	PromptEvalCount    int64         `json:"prompt_eval_count,omitempty"`    //9
	PromptEvalDuration time.Duration `json:"prompt_eval_duration,omitempty"` //195979000
	EvalCount          int64         `json:"eval_count,omitempty"`           //97
	EvalDuration       time.Duration `json:"eval_duration,omitempty"`        //16281091000
	Context            []int         `json:"-"`                              // context: [32010,3750,338,278,...,6575,4366,1213,13,29913]
}

func (of *ollamaFlow) Process(recs []engine.Record) ([]engine.Record, error) {
	ret := make([]engine.Record, 0, len(recs))
	for _, rec := range recs {
		genReq := &GenerateRequest{Model: of.model, Stream: of.stream}
		genReq.Prompt = "Where is the capital city of Australia?"

		if promptField := rec.Field("prompt"); promptField != nil && !promptField.IsNull() {
			if v, ok := promptField.Value.String(); ok {
				genReq.Prompt = v
			}
		}
		if imageField := rec.Field("image"); imageField != nil && !imageField.IsNull() {
			if imageField.Type() == engine.STRING {
				if v, ok := imageField.Value.String(); ok {
					genReq.Images = []string{v}
				}
			} else if imageField.Type() == engine.BINARY {
				if bv, ok := imageField.Value.Bytes(); ok {
					genReq.Images = []string{base64.StdEncoding.EncodeToString(bv)}
				}
			}
		}
		if modelField := rec.Field("model"); modelField != nil && !modelField.IsNull() {
			if v, ok := modelField.Value.String(); ok {
				genReq.Model = v
			}
		}
		if streamField := rec.Field("stream"); streamField != nil && !streamField.IsNull() {
			if v, ok := streamField.Value.Bool(); ok {
				genReq.Stream = v
			}
		}
		if formatField := rec.Field("format"); formatField != nil && !formatField.IsNull() {
			if v, ok := formatField.Value.String(); ok {
				genReq.Format = v
			}
		}
		of.ctx.LogDebug("flows.ollama", "model", genReq.Model, "prompt", genReq.Prompt, "stream", genReq.Stream, "images", len(genReq.Images))

		rt, err := of.process0(genReq)
		if err != nil {
			return nil, err
		}
		reFields := make([]*engine.Field, 0, len(rec.Fields()))
		for _, f := range rec.Fields() {
			if f.Name == "prompt" || f.Name == "image" || f.Name == "model" || f.Name == "stream" || f.Name == "format" {
				continue
			}
			reFields = append(reFields, f)
		}
		for i, rtRec := range rt {
			rt[i] = rtRec.Append(reFields...)
		}

		ret = append(ret, rt...)
	}
	return ret, nil
}

func (of *ollamaFlow) process0(genReq *GenerateRequest) ([]engine.Record, error) {
	uri, err := url.JoinPath(of.addr, "/api/generate")
	if err != nil {
		return nil, err
	}
	successCode := 200
	reqBody, err := json.Marshal(genReq)
	if err != nil {
		return nil, err
	}
	rsp, err := of.client.Post(uri, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != successCode {
		of.ctx.LogWarn("flows.ollama", "status", rsp.StatusCode, "body", string(body))
		rec := engine.NewRecord(engine.NewIntField("status", int64(rsp.StatusCode)))
		return []engine.Record{rec}, nil
	}

	if contentType := rsp.Header.Get("Content-Type"); strings.Contains(contentType, "application/json") {
		genRsp := GenerateResponse{}
		if err := json.Unmarshal(body, &genRsp); err != nil {
			of.ctx.LogWarn("flows.ollama", "status", rsp.StatusCode, "unmarshal error", err.Error())
			return nil, err
		}
		rec := of.parseResponse(&genRsp)
		return []engine.Record{rec}, nil
	} else if strings.Contains(contentType, "application/x-ndjson") {
		dec := json.NewDecoder(bytes.NewReader(body))
		ret := make([]engine.Record, 0, 30)
		for {
			genRsp := GenerateResponse{}
			if err := dec.Decode(&genRsp); err != nil {
				if err == io.EOF {
					break
				}
				of.ctx.LogWarn("flows.ollama", "status", rsp.StatusCode, "decode error", err.Error())
				return nil, err
			}
			rec := of.parseResponse(&genRsp)
			ret = append(ret, rec)
		}
		return ret, nil
	} else {
		of.ctx.LogWarn("flows.ollama", "status", rsp.StatusCode, "unsupported content-type", contentType)
		return nil, nil
	}
}

func (of *ollamaFlow) parseResponse(genRsp *GenerateResponse) engine.Record {
	rec := engine.NewRecord()
	if genRsp.Error != "" {
		rec = rec.Append(engine.NewStringField("error", genRsp.Error))
	} else {
		rec = rec.Append(
			engine.NewStringField("response", genRsp.Response),
			engine.NewTimeField("created_at", genRsp.CreateAt),
			engine.NewBoolField("done", genRsp.Done),
			engine.NewStringField("done_reason", genRsp.DoneReason),
			engine.NewIntField("total_duration", int64(genRsp.TotalDuration)),
		)
	}
	return rec
}
