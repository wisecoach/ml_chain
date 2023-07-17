package role

import (
	"fmt"
	"github.com/coniks-sys/coniks-go/crypto/vrf"
	"github.com/wisecoach/ml_chain/comm/comm"
	"reflect"
)

// Selector used for select the verifiers from all trainers
type Selector struct {
	verifierSk vrf.PrivateKey // Used for the verifier
	verifierPk vrf.PublicKey
}

// init the selector
func (rs *Selector) init() {
	var err error
	rs.verifierSk, err = vrf.GenerateKey(nil)
	if err != nil {
		fmt.Println("Error! Could not generate secret key for roles")
	}

	rs.verifierPk, _ = rs.verifierSk.Public()
}

// SelectVerifiers
//
//	@Description: 			select verifiers from trainers
//	@param trainers 		the pk list of trainers
//	@param input			the input of vrf
//	@param numRequested		the number of verifiers to be selected
//	@return []*comm.RemotePeer		the pk list of verifiers
//	@return []byte			the output of vrf
//	@return []byte			the proof of output
func (rs *Selector) SelectVerifiers(trainers []*comm.RemotePeer, input []byte, numRequested int) *SelectionResult {
	// we should snapshot the trainers
	currentTrainers := make([]*comm.RemotePeer, len(trainers))
	copy(currentTrainers, trainers)

	// generate the vrf prove and select the verifiers according to every two bytes of vrf prove
	prove, proof := rs.verifierSk.Prove(input)
	verifiers := rs.selectVerifierByProve(currentTrainers, prove, numRequested)

	return &SelectionResult{
		candidates: currentTrainers,
		winners:    verifiers,
		input:      input,
		prove:      prove,
		proof:      proof,
		pk:         rs.verifierPk,
	}
}

func (rs *Selector) selectVerifierByProve(trainers []*comm.RemotePeer, prove []byte, numRequested int) []*comm.RemotePeer {
	verifiers := make([]*comm.RemotePeer, numRequested)
	nodeMap := make(map[int]bool)

	i := 0
	for len(verifiers) < numRequested {
		index := (int(prove[i])*256 + int(prove[i+1])) % len(trainers)

		if !nodeMap[index] {
			verifiers = append(verifiers, trainers[index])
			nodeMap[index] = true
		}
		i++
	}

	return verifiers
}

// Verify
//
//	@Description:	verify the vrf random result
//	@return bool	if valid
func (rs *Selector) Verify(proof *SelectionResult) bool {
	if !proof.pk.Verify(proof.input, proof.prove, proof.proof) {
		return false
	}
	gotted := rs.selectVerifierByProve(proof.candidates, proof.input, len(proof.winners))
	return reflect.DeepEqual(proof.winners, gotted)
}

// SelectionResult
// @Description: used for other peer to verify our random selection
type SelectionResult struct {
	candidates []*comm.RemotePeer // TODO 该字段考虑删除
	winners    []*comm.RemotePeer
	input      []byte
	prove      []byte
	proof      []byte
	pk         vrf.PublicKey
}
