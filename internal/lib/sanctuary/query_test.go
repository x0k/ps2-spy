package sanctuary

import (
	"testing"
)

func TestNewQuery(t *testing.T) {
	type args struct {
		queryType  string
		namespace  string
		collection string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			args: args{GetQuery, Ns_ps2, WorldPopulationCollection},
			want: "get/ps2/world_population?c:censusJSON=false",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQuery(tt.args.queryType, tt.args.namespace, tt.args.collection).String(); got != tt.want {
				t.Errorf("NewQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
