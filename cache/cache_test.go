package cache

import "testing"

func Test_cache_CreateFilePath(t *testing.T) {
	tests := []struct {
		name      string
		baseDir   string
		baseNames []string
		want      string
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			"Should return nested safe path with extension",
			"base",
			[]string{"foo", "../bar", "myfile.ext"},
			"base/foo/-bar/myfile.ext",
			false,
		},
		{
			"Should still be contained within dir",
			"base",
			[]string{"/foo", "../bar", "myfile.ext"},
			"base/-foo/-bar/myfile.ext",
			false,
		},
		{
			"Should not have multiple 'escape'-charactes in a row",
			"base",
			[]string{"/../foo", "../../////bar../", "myfile.ext"},
			"base/-foo/-bar-/myfile.ext",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cache{
				baseDir: tt.baseDir,
			}
			got, err := c.CreateFilePath(tt.baseNames)
			if (err != nil) != tt.wantErr {
				t.Errorf("cache.CreateFilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("cache.CreateFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
