package client

import (
	"reflect"
	"testing"
)

func Test_splitPath(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Absolute path",
			args: args{
				dir: "/foo/bar/baz",
			},
			want: []string{"/foo", "/foo/bar"},
		},
		{
			name: "Relative path",
			args: args{
				dir: "path/to/some/file",
			},
			want: []string{"path", "path/to", "path/to/some"},
		},
		{
			name: "Relative with dot",
			args: args{
				dir: "./path/to/some/file",
			},
			want: []string{"path", "path/to", "path/to/some"},
		},
		{
			name: "Root",
			args: args{
				dir: "/",
			},
			want: []string{},
		},
		{
			name: "Trailing slash",
			args: args{
				dir: "/foo/bar/baz/",
			},
			want: []string{"/foo", "/foo/bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitPath(tt.args.dir); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
