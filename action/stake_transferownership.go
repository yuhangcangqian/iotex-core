// Copyright (c) 2020 IoTeX Foundation
// This source code is provided 'as is' and no warranties are given as to title or non-infringement, merchantability
// or fitness for purpose and, to the extent permitted by law, all liability for your use of the code is disclaimed.
// This source code is governed by Apache License 2.0 that can be found in the LICENSE file.

package action

import (
	"bytes"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/iotexproject/iotex-core/v2/pkg/util/byteutil"
)

const (
	_transferStakeInterfaceABI = `[
		{
			"inputs": [
				{
					"internalType": "address",
					"name": "voterAddress",
					"type": "address"
				},
				{
					"internalType": "uint64",
					"name": "bucketIndex",
					"type": "uint64"
				},
				{
					"internalType": "uint8[]",
					"name": "data",
					"type": "uint8[]"
				}
			],
			"name": "transferStake",
			"outputs": [],
			"stateMutability": "nonpayable",
			"type": "function"
		}
	]`
)

var (
	// _transferStakeMethod is the interface of the abi encoding of stake action
	_transferStakeMethod abi.Method
	_                    EthCompatibleAction = (*TransferStake)(nil)
)

// TransferStake defines the action of transfering stake ownership ts the other
type TransferStake struct {
	stake_common
	voterAddress address.Address
	bucketIndex  uint64
	payload      []byte
}

func init() {
	transferStakeInterface, err := abi.JSON(strings.NewReader(_transferStakeInterfaceABI))
	if err != nil {
		panic(err)
	}
	var ok bool
	_transferStakeMethod, ok = transferStakeInterface.Methods["transferStake"]
	if !ok {
		panic("fail to load the method")
	}
}

// NewTransferStake returns a TransferStake instance
func NewTransferStake(
	voterAddress string,
	bucketIndex uint64,
	payload []byte,
) (*TransferStake, error) {
	voterAddr, err := address.FromString(voterAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load address from string")
	}
	return &TransferStake{
		voterAddress: voterAddr,
		bucketIndex:  bucketIndex,
		payload:      payload,
	}, nil
}

// VoterAddress returns the address of recipient
func (ts *TransferStake) VoterAddress() address.Address { return ts.voterAddress }

// BucketIndex returns bucket index
func (ts *TransferStake) BucketIndex() uint64 { return ts.bucketIndex }

// Payload returns the payload bytes
func (ts *TransferStake) Payload() []byte { return ts.payload }

// Serialize returns a raw byte stream of the transfer stake action struct
func (ts *TransferStake) Serialize() []byte {
	return byteutil.Must(proto.Marshal(ts.Proto()))
}

func (act *TransferStake) FillAction(core *iotextypes.ActionCore) {
	core.Action = &iotextypes.ActionCore_StakeTransferOwnership{StakeTransferOwnership: act.Proto()}
}

// Proto converts transfer stake to protobuf
func (ts *TransferStake) Proto() *iotextypes.StakeTransferOwnership {
	act := &iotextypes.StakeTransferOwnership{
		VoterAddress: ts.voterAddress.String(),
		BucketIndex:  ts.bucketIndex,
		Payload:      ts.payload,
	}

	return act
}

// LoadProto loads transfer stake protobuf
func (ts *TransferStake) LoadProto(pbAct *iotextypes.StakeTransferOwnership) error {
	if pbAct == nil {
		return ErrNilProto
	}
	voterAddress, err := address.FromString(pbAct.GetVoterAddress())
	if err != nil {
		return errors.Wrap(err, "failed to load address from string")
	}
	ts.voterAddress = voterAddress
	ts.bucketIndex = pbAct.GetBucketIndex()
	ts.payload = pbAct.GetPayload()
	return nil
}

// IntrinsicGas returns the intrinsic gas of a TransferStake
func (ts *TransferStake) IntrinsicGas() (uint64, error) {
	payloadSize := uint64(len(ts.Payload()))
	return CalculateIntrinsicGas(MoveStakeBaseIntrinsicGas, MoveStakePayloadGas, payloadSize)
}

func (ts *TransferStake) SanityCheck() error {
	return nil
}

// EthData returns the ABI-encoded data for converting to eth tx
func (ts *TransferStake) EthData() ([]byte, error) {
	voterEthAddr := common.BytesToAddress(ts.voterAddress.Bytes())
	data, err := _transferStakeMethod.Inputs.Pack(voterEthAddr, ts.bucketIndex, ts.payload)
	if err != nil {
		return nil, err
	}
	return append(_transferStakeMethod.ID, data...), nil
}

// NewTransferStakeFromABIBinary decodes data into TransferStake action
func NewTransferStakeFromABIBinary(data []byte) (*TransferStake, error) {
	var (
		paramsMap = map[string]interface{}{}
		ok        bool
		err       error
		ts        TransferStake
	)
	// sanity check
	if len(data) <= 4 || !bytes.Equal(_transferStakeMethod.ID, data[:4]) {
		return nil, errDecodeFailure
	}
	if err := _transferStakeMethod.Inputs.UnpackIntoMap(paramsMap, data[4:]); err != nil {
		return nil, err
	}
	if ts.voterAddress, err = ethAddrToNativeAddr(paramsMap["voterAddress"]); err != nil {
		return nil, err
	}
	if ts.bucketIndex, ok = paramsMap["bucketIndex"].(uint64); !ok {
		return nil, errDecodeFailure
	}
	if ts.payload, ok = paramsMap["data"].([]byte); !ok {
		return nil, errDecodeFailure
	}
	return &ts, nil
}
