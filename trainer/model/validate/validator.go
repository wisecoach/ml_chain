package validate

import "github.com/wisecoach/ml_chain/proto"

type Validator interface {
	Validate(weight *proto.LocalityWeight)
}
