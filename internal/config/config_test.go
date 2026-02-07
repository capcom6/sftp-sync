package config_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/capcom6/sftp-sync/internal/config"
)

func TestParse(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name    string
		args    args
		want    config.Config
		wantErr bool
	}{
		{
			name: "Empty",
			args: args{},
			want: config.Config{
				WatchPath:    "",
				ExcludePaths: []string{},
				Dest:         "",
			},
			wantErr: true,
		},
		{
			name: "Source and dest",
			args: args{
				args: []string{"--dest", "dest", "path"},
			},
			want: config.Config{
				WatchPath:    "path",
				ExcludePaths: []string{},
				Dest:         "dest",
			},
			wantErr: false,
		},
		{
			name: "Single exclude",
			args: args{
				args: []string{"--dest", "dest", "--exclude", "ex1", "path"},
			},
			want: config.Config{
				WatchPath:    "path",
				ExcludePaths: []string{"ex1"},
				Dest:         "dest",
			},
			wantErr: false,
		},
		{
			name: "Multiple excludes",
			args: args{
				args: []string{"--dest", "dest", "--exclude", "ex1", "--exclude", "ex2", "path"},
			},
			want: config.Config{
				WatchPath:    "path",
				ExcludePaths: []string{"ex1", "ex2"},
				Dest:         "dest",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.Parse(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !errors.Is(err, config.ErrValidationFailed) {
				t.Errorf("Parse() error = %v, want %v", err, config.ErrValidationFailed)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %+#v, want %+#v", got, tt.want)
			}
		})
	}
}
