package store

import (
	"reflect"
	"testing"

	"go.uber.org/zap"
)

func TestNewStore(t *testing.T) {
	type args struct {
		dsn string
		lg  *zap.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    Storage
		wantErr bool
	}{
		{
			name: "Storage must be return error",
			args: args{
				dsn: "test",
				lg:  zap.NewNop(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewStore(tt.args.dsn, tt.args.lg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStore() = %v, want %v", got, tt.want)
			}
		})
	}
}
