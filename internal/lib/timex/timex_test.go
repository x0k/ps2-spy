package timex_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/timex"
)

func TestShiftDate(t *testing.T) {
	type args struct {
		weekday time.Weekday
		t1      time.Duration
		offset  time.Duration
	}
	tests := []struct {
		name  string
		args  args
		want  time.Weekday
		want1 time.Duration
	}{
		{
			name: "simple1",
			args: args{
				weekday: time.Monday,
				t1:      2 * time.Hour,
				offset:  -3 * time.Hour,
			},
			want:  time.Sunday,
			want1: 23 * time.Hour,
		},
		{
			name: "simple2",
			args: args{
				weekday: time.Sunday,
				t1:      23 * time.Hour,
				offset:  2 * time.Hour,
			},
			want:  time.Monday,
			want1: 1 * time.Hour,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := timex.NormalizeDate(tt.args.weekday, tt.args.t1+tt.args.offset)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShiftDate() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ShiftDate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
