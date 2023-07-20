package role

import (
	"fmt"
	"github.com/coniks-sys/coniks-go/crypto/vrf"
	"github.com/wisecoach/ml_chain/comm/comm"
	"github.com/wisecoach/ml_chain/proto"
	"reflect"
)

// Selector used for select the validators from all trainers
type Selector struct {
	validatorsk vrf.PrivateKey // Used for the validator
	validatorPk vrf.PublicKey
}

// init the selector
func (rs *Selector) init() {
	var err error
	rs.validatorsk, err = vrf.GenerateKey(nil)
	if err != nil {
		fmt.Println("Error! Could not generate secret key for roles")
	}

	rs.validatorPk, _ = rs.validatorsk.Public()
}

// SelectValidators
//
//	@Description: 			select validators from trainers
//	@param trainers 		the pk list of trainers
//	@param input			the input of vrf
//	@param numRequested		the number of validators to be selected
//	@return []*comm.RemotePeer		the pk list of validators
//	@return []byte			the output of vrf
//	@return []byte			the proof of output
func (rs *Selector) SelectValidators(trainers []*comm.RemotePeer, input []byte, numRequested int) *proto.SelectionResult {
	// we should snapshot the trainers
	currentTrainers := make([]*comm.RemotePeer, len(trainers))
	copy(currentTrainers, trainers)

	// generate the vrf prove and select the validators according to every two bytes of vrf prove
	prove, proof := rs.validatorsk.Prove(input)
	validators := rs.selectValidatorByProve(currentTrainers, prove, numRequested)

	return &proto.SelectionResult{
		Candidates: currentTrainers,
		Winners:    validators,
		Input:      input,
		Prove:      prove,
		Proof:      proof,
		Pk:         rs.validatorPk,
	}
}

func (rs *Selector) selectValidatorByProve(trainers []*comm.RemotePeer, prove []byte, numRequested int) []*comm.RemotePeer {
	validators := make([]*comm.RemotePeer, numRequested)
	nodeMap := make(map[int]bool)

	i := 0
	for len(validators) < numRequested {
		index := (int(prove[i])*256 + int(prove[i+1])) % len(trainers)

		if !nodeMap[index] {
			validators = append(validators, trainers[index])
			nodeMap[index] = true
		}
		i++
	}

	return validators
}

// Verify
//
//	@Description:	verify the vrf random result
//	@return bool	if valid
func (rs *Selector) Verify(proof *proto.SelectionResult) bool {
	if !proof.Pk.Verify(proof.Input, proof.Prove, proof.Proof) {
		return false
	}
	gotted := rs.selectValidatorByProve(proof.Candidates, proof.Input, len(proof.Winners))
	return reflect.DeepEqual(proof.Winners, gotted)
}
