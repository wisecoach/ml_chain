package trainer

import (
	"fmt"
	"github.com/coniks-sys/coniks-go/crypto/vrf"
	"reflect"
)

// RoleSelector used for select the verifiers from all trainers
type RoleSelector struct {
	verifierSk vrf.PrivateKey // Used for the verifier
	verifierPk vrf.PublicKey
}

// init the selector
func (rs *RoleSelector) init() {
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
//	@return [][]byte		the pk list of verifiers
//	@return []byte			the output of vrf
//	@return []byte			the proof of output
func (rs *RoleSelector) SelectVerifiers(trainers [][]byte, input []byte, numRequested int) ([][]byte, *SelectionProof) {
	// we should snapshot the trainers
	currentTrainers := make([][]byte, len(trainers))
	copy(currentTrainers, trainers)

	// generate the vrf prove and select the verifiers according to every two bytes of vrf prove
	prove, proof := rs.verifierSk.Prove(input)
	verifiers := rs.selectVerifierByProve(currentTrainers, prove, numRequested)

	return trainers[:numRequested], &SelectionProof{
		candidates: currentTrainers,
		winners:    verifiers,
		input:      input,
		prove:      prove,
		proof:      proof,
		pk:         rs.verifierPk,
	}
}

func (rs *RoleSelector) selectVerifierByProve(trainers [][]byte, prove []byte, numRequested int) [][]byte {
	verifiers := make([][]byte, numRequested)
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
func (rs *RoleSelector) Verify(proof *SelectionProof) bool {
	if !proof.pk.Verify(proof.input, proof.prove, proof.proof) {
		return false
	}
	gotted := rs.selectVerifierByProve(proof.candidates, proof.input, len(proof.winners))
	return reflect.DeepEqual(proof.winners, gotted)
}

// SelectionProof
// @Description: used for other peer to verify our random selection
type SelectionProof struct {
	candidates [][]byte
	winners    [][]byte
	input      []byte
	prove      []byte
	proof      []byte
	pk         vrf.PublicKey
}
