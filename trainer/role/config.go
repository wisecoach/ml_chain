package role

import "github.com/wisecoach/ml_chain/comm/comm"

type Config struct {
	Self            *comm.RemotePeer
	TaskManagerList [][]byte
}
