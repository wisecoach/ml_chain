package proto

import (
	"github.com/coniks-sys/coniks-go/crypto/vrf"
)

// SelectionResult
// @Description: used for other peer to verify our random selection
type SelectionResult struct {
	Candidates []*RemotePeer // TODO 该字段考虑删除，但是得考虑顺序不一致、节点发现不一致、不包含self等原因导致无法验证的问题
	Winners    []*RemotePeer
	Input      []byte
	Prove      []byte
	Proof      []byte
	Pk         vrf.PublicKey
}
