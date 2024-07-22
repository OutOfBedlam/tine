package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
}

func (of *ollamaFlow) Open() error {
	of.addr = of.ctx.Config().GetString("address", "")
	if !strings.HasPrefix(of.addr, "http://") && strings.HasPrefix(of.addr, "https://") {
		return fmt.Errorf("invalid address: %s", of.addr)
	}
	of.model = of.ctx.Config().GetString("model", "phi3")
	timeout := of.ctx.Config().GetDuration("timeout", 5*time.Second)

	of.ctx.LogDebug("flows.ollama", "address", of.addr, "timeout", timeout)

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
		rt, err := of.process0(rec)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt...)
	}
	return ret, nil
}

func (of *ollamaFlow) process0(rec engine.Record) ([]engine.Record, error) {
	uri, err := url.JoinPath(of.addr, "/api/generate")
	if err != nil {
		return nil, err
	}
	successCode := 200
	genReq := GenerateRequest{
		Model:  "phi3",
		Prompt: "Where is the capital city of Australia?",
		//Format: "json",
		Stream: true,
	}
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
		return nil, nil
	}

	if contentType := rsp.Header.Get("Content-Type"); strings.Contains(contentType, "application/json") {
		genRsp := GenerateResponse{}
		if err := json.Unmarshal(body, &genRsp); err != nil {
			of.ctx.LogWarn("flows.ollama", "status", rsp.StatusCode, "unmarshal error", err.Error())
			return nil, err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(genRsp)
		//of.ctx.LogInfo("flows.ollama", "response", genRsp)
		return []engine.Record{rec}, nil
	} else if strings.Contains(contentType, "application/x-ndjson") {
		dec := json.NewDecoder(bytes.NewReader(body))
		for {
			genRsp := GenerateResponse{}
			if err := dec.Decode(&genRsp); err != nil {
				if err == io.EOF {
					break
				}
				of.ctx.LogWarn("flows.ollama", "status", rsp.StatusCode, "decode error", err.Error())
				return nil, err
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(genRsp)
			//of.ctx.LogInfo("flows.ollama", "response", genRsp)
		}
		return []engine.Record{rec}, nil
	} else {
		of.ctx.LogWarn("flows.ollama", "status", rsp.StatusCode, "unsupported content-type", contentType)
		return nil, nil
	}
}
