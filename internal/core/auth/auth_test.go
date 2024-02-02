package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetJWT(t *testing.T) {
	type args struct {
		id    int
		login string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Get jwt token",
			args: args{
				id:    1,
				login: "test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetJWT(tt.args.id, tt.args.login)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestVerifyJWTandGetPayload(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Token shoul be correct",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetJWT(1, "test")
			if err != nil {
				assert.NoError(t, err)
			}

			_, err = VerifyJWTandGetPayload(*token)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyJWTandGetPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
