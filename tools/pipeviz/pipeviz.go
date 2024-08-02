package pipeviz

import (
	"fmt"
	"io"
	"os"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/emicklei/dot"
)

// Usage: tine graph ./tmp/ollama.toml -o - | dot -Tpng -o./tmp/graph.png
func Graph(writer io.Writer, args []string) error {
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
		g.Attr("fontname", "sans-serif,Arial,Helvetica")
		inlets, outlets, flows := []dot.Node{}, []dot.Node{}, []dot.Node{}
		nodeIdx := 0
		p.Walk(func(pipelineName, kind, name string, step any) {
			fullName := fmt.Sprintf("%s.%s", kind, name)
			nodeId := fmt.Sprintf("%d.%d", pIdx, nodeIdx)
			nodeIdx++
			node := g.Node(nodeId)
			node.Box()
			node.Label(fullName)
			node.Attr("weight", 0)

			if kind == "inlets" {
				theLastOfThisNode := node
				if handler, ok := step.(*engine.InletHandler); ok {
					subFlowIdx := 0
					handler.Walk(func(inletName string, subKind string, subStep string, subHandler any) {
						subNodeFullName := fmt.Sprintf("%s.%s.%s.%s", kind, inletName, subKind, subStep)
						subNodeId := fmt.Sprintf("%d.%d.%d", pIdx, nodeIdx, subFlowIdx)
						subFlowIdx++
						subNode := g.Node(subNodeId)
						subNode.Box()
						subNode.Label(subNodeFullName)
						subNode.Attr("weight", 0)
						theLastOfThisNode.Edge(subNode)
						theLastOfThisNode = subNode
					})
				}
				inlets = append(inlets, theLastOfThisNode)
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
