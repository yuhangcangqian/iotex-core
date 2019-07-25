// Copyright (c) 2019 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package bc

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotexproject/iotex-proto/golang/iotextypes"

	"github.com/iotexproject/iotex-core/ioctl/cmd/config"
	"github.com/iotexproject/iotex-core/ioctl/output"
)

// bcInfoCmd represents the bc info command
var bcInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get current block chain information",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		err := bcInfo()
		return output.PrintError(err)
	},
}

type infoMessage struct {
	Node string                `json:"node"`
	Info *iotextypes.ChainMeta `json:"info"`
}

func (m *infoMessage) String() string {
	if output.Format == "" {
		message := fmt.Sprintf("Blockchain Node: %s\n%s", m.Node, output.JSONString(m.Info))
		return message
	}
	return output.FormatString(output.Result, m)
}

// bcInfo get current information of block chain from server
func bcInfo() error {
	chainMeta, err := GetChainMeta()
	if err != nil {
		return output.NewError(0, "failed to get chain meta", err)
	}
	message := infoMessage{Node: config.ReadConfig.Endpoint, Info: chainMeta}
	fmt.Println(message.String())
	return nil
}
