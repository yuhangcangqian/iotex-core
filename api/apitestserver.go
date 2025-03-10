// Copyright (c) 2019 IoTeX Foundation
// This source code is provided 'as is' and no warranties are given as to title or non-infringement, merchantability
// or fitness for purpose and, to the extent permitted by law, all liability for your use of the code is disclaimed.
// This source code is governed by Apache License 2.0 that can be found in the LICENSE file.

package api

import (
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"google.golang.org/grpc"
)

// StreamBlocksServer defines the interface of a rpc stream server
type StreamBlocksServer interface {
	grpc.ServerStream
	Send(*iotexapi.StreamBlocksResponse) error
}
