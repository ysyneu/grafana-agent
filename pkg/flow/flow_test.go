package flow

import (
	"os"
	"testing"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/flow/internal/testcomponents"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/stretchr/testify/require"
)

var testFile = `
	testcomponents "tick" "ticker" {
		frequency = "1s"
	}

	testcomponents "passthrough" "static" {
		input = "hello, world!"
	}

	testcomponents "passthrough" "ticker" {
		input = testcomponents.tick.ticker.tick_time
	}

	testcomponents "passthrough" "forwarded" {
		input = testcomponents.passthrough.ticker.output
	}
`

func TestController_LoadFile_Evaluation(t *testing.T) {
	ctrl, _ := newFlow(testOptions(t))

	// Use testFile from graph_builder_test.go.
	f, diags := ReadFile(t.Name(), []byte(testFile))
	require.False(t, diags.HasErrors())
	require.NotNil(t, f)

	err := ctrl.LoadFile(f)
	require.NoError(t, err)
	require.Len(t, ctrl.loader.Components(), 4)

	// Check the inputs and outputs of things that should be immediately resolved
	// without having to run the components.
	in, out := getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.static")
	require.Equal(t, "hello, world!", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "hello, world!", out.(testcomponents.PassthroughExports).Output)
}

func getFields(t *testing.T, g *dag.Graph, nodeID string) (component.Arguments, component.Exports) {
	t.Helper()

	n := g.GetByID(nodeID)
	require.NotNil(t, n, "couldn't find node %q in graph", nodeID)

	uc := n.(*controller.ComponentNode)
	return uc.Arguments(), uc.Exports()
}

func testOptions(t *testing.T) Options {
	t.Helper()

	l, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)

	return Options{
		Logger:   l,
		DataPath: t.TempDir(),
	}
}
