package worker

import (
	"testing"

	"github.com/dedpnd/unifier/internal/models"
	"github.com/stretchr/testify/assert"
)

func Test_extraProcess(t *testing.T) {
	type args struct {
		cfgExtraProcess []models.ExtraProcess
		uEvent          *map[string]interface{}
	}
	type want struct {
		key   string
		value string
	}

	tests := []struct {
		name string
		args args
		want
		wantErr bool
	}{
		{
			name: "Func __if must write value",
			args: args{
				cfgExtraProcess: []models.ExtraProcess{{
					Func: "__if",
					Args: "testFrom, 123, 321",
					To:   "testTo",
				}},
				uEvent: &map[string]interface{}{
					"testFrom": "123",
				},
			},
			want: want{
				key:   "testTo",
				value: "321",
			},
			wantErr: false,
		},
		{
			name: "Func __stringConstant must write value",
			args: args{
				cfgExtraProcess: []models.ExtraProcess{{
					Func: "__stringConstant",
					Args: "test",
					To:   "testTo",
				}},
				uEvent: &map[string]interface{}{
					"testFrom": "123",
				},
			},
			want: want{
				key:   "testTo",
				value: "test",
			},
			wantErr: false,
		},
		{
			name: "Func __testFunc should return an error",
			args: args{
				cfgExtraProcess: []models.ExtraProcess{{
					Func: "__testFunc",
					Args: "test",
					To:   "testTo",
				}},
				uEvent: &map[string]interface{}{
					"testFrom": "123",
				},
			},
			want: want{
				key:   "testFrom",
				value: "123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := extraProcess(tt.args.cfgExtraProcess, tt.args.uEvent); (err != nil) != tt.wantErr {
				t.Errorf("extraProcess() error = %v, wantErr %v", err, tt.wantErr)
			}

			v, ok := (*tt.args.uEvent)[tt.want.key]
			if !ok {
				t.Errorf("field = %v must be exist", tt.want.key)
			}

			assert.Equal(t, tt.want.value, v)
		})
	}
}

func Test_unificationFields(t *testing.T) {
	type args struct {
		event      map[string]interface{}
		cfgUnifier []models.Unifier
		uEvent     *map[string]interface{}
	}
	type want struct {
		key   string
		value interface{}
	}

	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "Field must be string",
			args: args{
				event: map[string]interface{}{
					"testFrom": "qwerty",
				},
				cfgUnifier: []models.Unifier{{
					Name:       "testString",
					Type:       "string",
					Expression: "testFrom",
				}},
				uEvent: &map[string]interface{}{},
			},
			want: want{
				key:   "testString",
				value: "qwerty",
			},
			wantErr: false,
		},
		{
			name: "Field must be int",
			args: args{
				event: map[string]interface{}{
					"testFrom": "123",
				},
				cfgUnifier: []models.Unifier{{
					Name:       "testInt",
					Type:       "int",
					Expression: "testFrom",
				}},
				uEvent: &map[string]interface{}{},
			},
			want: want{
				key:   "testInt",
				value: 123,
			},
			wantErr: false,
		},
		{
			name: "Field must be timestamp",
			args: args{
				event: map[string]interface{}{
					"testFrom": "2023-07-13T13:47:43+00:00",
				},
				cfgUnifier: []models.Unifier{{
					Name:       "testTime",
					Type:       "timestamp",
					Expression: "testFrom",
				}},
				uEvent: &map[string]interface{}{},
			},
			want: want{
				key:   "testTime",
				value: "2023-07-13 13:47:43",
			},
			wantErr: false,
		},
		{
			name: "Should return an error",
			args: args{
				event: map[string]interface{}{
					"testFrom": "123",
				},
				cfgUnifier: []models.Unifier{{
					Name:       "test",
					Type:       "testType",
					Expression: "testFrom",
				}},
				uEvent: &map[string]interface{}{
					"testFrom": "123",
				},
			},
			want: want{
				key:   "testFrom",
				value: "123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := unificationFields(tt.args.event, tt.args.cfgUnifier, tt.args.uEvent); (err != nil) != tt.wantErr {
				t.Errorf("unificationFields() error = %v, wantErr %v", err, tt.wantErr)
			}

			v, ok := (*tt.args.uEvent)[tt.want.key]
			if !ok {
				t.Errorf("field = %v must be exist", tt.want.key)
			}

			assert.Equal(t, tt.want.value, v)
		})
	}
}

func Test_calculateHash(t *testing.T) {
	type args struct {
		event      map[string]interface{}
		cfgEntHash []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Get hash must be correct",
			args: args{
				event: map[string]interface{}{
					"testString": "qwerty",
					"testInt":    123,
				},
				cfgEntHash: []string{"testString", "testInt"},
			},
			want:    "d8578edf8458ce06fbc5bb76a58c5ca4",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateHash(tt.args.event, tt.args.cfgEntHash)
			if got != tt.want {
				t.Errorf("calculateHash() = %v, want %v", got, tt.want)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
