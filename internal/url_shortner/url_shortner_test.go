package urlshortner

import (
	"testing"
)

func TestCalculateMinimumLength(t *testing.T) {
	type args struct {
		n int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "test estimate 1mil length",
			args: args{
				n: 1_000_000,
			},
			want: 4,
		},
		{
			name: "test estimate 10 length",
			args: args{
				n: 10,
			},
			want: 4,
		},
		{
			name: "test estimate 900mil length",
			args: args{
				n: 916_132_832,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateMinimumLength(tt.args.n); got != tt.want {
				t.Errorf("EstimateLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
