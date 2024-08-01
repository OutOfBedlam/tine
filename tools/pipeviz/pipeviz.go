package pipeviz

import (
	"fmt"
	"os"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/emicklei/dot"
)

// Usage: tine graph ./tmp/ollama.toml -o - | dot -Tpng -o./tmp/graph.png
func Graph(args []string, output string) error {
	pipelineConfigs := []string{}
	for _, arg := range args {
		if _, err := os.Stat(arg); err != nil {
			return fmt.Errorf("pipeline file not found: %s", arg)
		}
		pipelineConfigs = append(pipelineConfigs, arg)
	}

	if len(pipelineConfigs) == 0 {
		return fmt.Errorf("no pipeline file specified")
	}

	pipelines := make([]*engine.Pipeline, 0, len(pipelineConfigs))
	for _, pc := range pipelineConfigs {
		p, err := engine.New(engine.WithConfigFile(pc))
		if err != nil {
			return fmt.Errorf("failed to parse pipeline file: %w", err)
		}
		pipelines = append(pipelines, p)
	}
	for _, p := range pipelines {
		if err := p.Build(); err != nil {
			return err
		}
	}
	gg := dot.NewGraph(dot.Directed)
	for pIdx, p := range pipelines {
		g := gg.Subgraph(p.Name)
		g.Attr("fontname", "Helvetica,Arial,sans-serif")
		inlets, outlets, flows := []dot.Node{}, []dot.Node{}, []dot.Node{}
		nodeIdx := 0
		p.Walk(func(pipelineName, kind, step string) {
			fullName := fmt.Sprintf("%s.%s", kind, step)
			nodeId := fmt.Sprintf("%d.%d", pIdx, nodeIdx)
			nodeIdx++
			node := g.Node(nodeId)
			node.Box()
			node.Label(fmt.Sprintf("[[%s]]", fullName))
			node.Attr("weight", 0)

			if kind == "inlets" {
				inlets = append(inlets, node)
			} else if kind == "outlets" {
				outlets = append(outlets, node)
			} else if kind == "flows" {
				flows = append(flows, node)
			}
		})
		if len(flows) > 0 {
			for _, inlet := range inlets {
				inlet.Edge(flows[0])
			}
			for i, flow := range flows {
				if i == 0 {
					continue
				}
				flows[i-1].Edge(flow)
			}
			for _, outlet := range outlets {
				flows[len(flows)-1].Edge(outlet)
			}
		} else if len(outlets) > 0 {
			for _, inlet := range inlets {
				inlet.Edge(outlets[0])
			}
		}
	}

	writer := os.Stdout
	if output != "-" {
		if w, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
			return err
		} else {
			writer = w
		}
	}
	_, err := writer.Write([]byte(gg.String()))
	return err
}

// strict digraph {
// 	"example.toml:flows.fan-in" [  weight=0 ];
// 	"example.toml:flows.fan-in" -> "example.toml:outlets.file" [  weight=0 ];
// 	"example.toml:outlets.file" [  weight=0 ];
// 	"example.toml:inlets.cpu" [  weight=0 ];
// 	"example.toml:inlets.cpu" -> "example.toml:flows.fan-in" [  weight=0 ];
// }
