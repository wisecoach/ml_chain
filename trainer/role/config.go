package role

import "github.com/wisecoach/ml_chain/proto"

type Config struct {
	Self            *proto.RemotePeer
	TaskManagerList [][]byte
}
