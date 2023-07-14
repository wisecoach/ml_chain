package main

import (
	"github.com/coniks-sys/coniks-go/crypto/vrf"
	"reflect"
	"testing"
)

func TestVRF_getVRFRoles(t *testing.T) {
	type fields struct {
		rolesSk vrf.PrivateKey
		rolesPk vrf.PublicKey
		noiseSk vrf.PrivateKey
		noisePk vrf.PublicKey
	}
	type args struct {
		stakeMap     map[int]int
		input        []byte
		numVerifiers int
		numMiners    int
		totalNodes   int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []int
		want1  []int
		want2  []byte
		want3  []byte
	}{
		{
			name:   "",
			fields: fields{},
			args: args{
				stakeMap:     nil,
				input:        nil,
				numVerifiers: 0,
				numMiners:    0,
				totalNodes:   0,
			},
			want:  nil,
			want1: nil,
			want2: nil,
			want3: nil,
		},
		{
			name:   "",
			fields: fields{},
			args: args{
				stakeMap:     nil,
				input:        nil,
				numVerifiers: 0,
				numMiners:    0,
				totalNodes:   0,
			},
			want:  nil,
			want1: nil,
			want2: nil,
			want3: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			myvrf := &VRF{}
			myvrf.init()
			stakeMap := make(map[int]int)
			stakeMap[0] = 200
			stakeMap[1] = 200
			stakeMap[2] = 200
			got, got1, got2, got3 := myvrf.getVRFRoles(stakeMap, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}, 1, 1, 3)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVRFRoles() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getVRFRoles() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("getVRFRoles() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("getVRFRoles() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}
