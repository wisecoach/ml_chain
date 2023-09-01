package python

import (
	"github.com/wisecoach/ml_chain/proto"
	"net/http"
	"testing"
)

func Test_httpPythonClient_Init(t *testing.T) {
	type fields struct {
		TaskId     string
		ApiBaseUrl string
		httpClient *http.Client
	}
	type args struct {
		genesis *proto.TaskGenesis
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			fields: fields{
				TaskId:     "task1",
				ApiBaseUrl: "http://localhost:8888",
				httpClient: &http.Client{},
			},
			args: args{
				genesis: &proto.TaskGenesis{
					TaskId: "",
					ModelStructure: &proto.ModelStructure{
						Dataset:      "mnist",
						NumClasses:   10,
						Agent:        1,
						TrainerNum:   40,
						LearningRate: 0.01,
						Momentum:     0.5,
						Dp:           false,
						DpEpsilon:    0.4,
						DpEpsilon1:   0.4,
						DpDelta:      1e-5,
						DpClip:       300,
						BatchSize:    64,
						Round:        100,
						Lambda:       1,
					},
					ManagerList: [][]byte{{'a', 'b', 'c'}, {'c', 'd'}},
					InitWeight: &proto.Envelope[*proto.GlobalWeight]{
						Payload: &proto.GlobalWeight{
							Iteration:    0,
							WeightVector: make([]float32, 656080),
							Aggregator:   nil,
						},
						Signature: nil,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpPythonClient{
				TaskId:     tt.fields.TaskId,
				ApiBaseUrl: tt.fields.ApiBaseUrl,
				httpClient: tt.fields.httpClient,
			}
			if err := h.Init(tt.args.genesis); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
