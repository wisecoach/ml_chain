package role

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/coniks-sys/coniks-go/crypto/vrf"
	"github.com/wisecoach/ml_chain/proto"
	"reflect"
)

// Selector used for select the validators from all trainers
type Selector struct {
	validatorSk vrf.PrivateKey // Used for the validator
	validatorPk vrf.PublicKey
}

// init the selector
func (rs *Selector) init() {
	var err error
	rs.validatorSk, err = vrf.GenerateKey(nil)
	if err != nil {
		fmt.Println("Error! Could not generate secret key for roles")
	}

	rs.validatorPk, _ = rs.validatorSk.Public()
}

// SelectValidators
//
//	@Description: 			select validators from trainers
//	@param trainers 		the pk list of trainers
//	@param input			the input of vrf
//	@param numRequested		the number of validators to be selected
//	@return []*proto.RemotePeer		the pk list of validators
//	@return []byte			the output of vrf
//	@return []byte			the proof of output
func (rs *Selector) SelectValidators(trainers []*proto.RemotePeer, input []byte, numRequested int) *proto.SelectionResult {
	// we should snapshot the trainers
	currentTrainers := make([]*proto.RemotePeer, len(trainers))
	copy(currentTrainers, trainers)

	// generate the vrf prove and select the validators according to every two bytes of vrf prove
	prove, proof := rs.validatorSk.Prove(input)
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

func (rs *Selector) selectValidatorByProve(trainers []*proto.RemotePeer, prove []byte, numRequested int) []*proto.RemotePeer {
	validators := make([]*proto.RemotePeer, 0)
	nodeMap := make(map[int]bool)

	i := 0
	for len(validators) < numRequested {

		if i >= (len(prove) - 1) {
			hash := sha256.Sum256(prove)
			prove = hash[:]
			i = 0
		}

		index := (int(prove[i])*256 + int(prove[i+1])) % len(trainers)

		if !nodeMap[index] {
			validators = append(validators, trainers[index])
			nodeMap[index] = true
		}
		i++
	}

	return validators
}

func (rs *Selector) VerifyValidatorSelection(proof *proto.SelectionResult) error {
	if !proof.Pk.Verify(proof.Input, proof.Prove, proof.Proof) {
		return errors.New("the vrf result is not valid")
	}
	gotted := rs.selectValidatorByProve(proof.Candidates, proof.Input, len(proof.Winners))
	if reflect.DeepEqual(proof.Winners, gotted) {
		return nil
	} else {
		return errors.New("the vrf result is not valid")
	}
}
