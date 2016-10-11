package intg

import (
	"fmt"

	"github.com/markdaws/gohome"
	"github.com/markdaws/gohome/cmd"
)

// Returns a cmd.Builder given the builder ID e.g. "fluxwifi"
func CmdBuilderFromID(system *gohome.System, ID string) (cmd.Builder, error) {
	switch ID {
	case "belkin-wemo-insight":
		return &belkinCmdBuilder{system}, nil
	case "fluxwifi":
		return &fluxwifiCmdBuilder{system}, nil
	case "tcp600gwb":
		return &tcp600gwbCmdBuilder{system}, nil
	case "l-bdgpro2-wh":
		return &lbdgpro2whCmdBuilder{System: system}, nil
	default:
		return nil, fmt.Errorf("unsupported command builder ID %s", ID)
	}
}
