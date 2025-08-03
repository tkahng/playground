package security

import (
	"testing"
)

func TestEstimateLength(t *testing.T) {
	type args struct {
		n            int64
		alphabetSize int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "test estimate 1mil length",
			args: args{
				n:            1_000_000,
				alphabetSize: 62,
			},
			want: 4,
		},
		{
			name: "test estimate 10 length",
			args: args{
				n:            10,
				alphabetSize: 62,
			},
			want: 1,
		},
		{
			name: "test estimate 900mil length",
			args: args{
				n:            916_132_832,
				alphabetSize: 62,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EstimateLength(tt.args.n, tt.args.alphabetSize); got != tt.want {
				t.Errorf("EstimateLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
