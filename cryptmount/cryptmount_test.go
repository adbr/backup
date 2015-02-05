// 2015-01-28 Adam Bryt

package main

import (
	"testing"
)

func TestDiskName(t *testing.T) {
	getDiskNames = func() (string, error) {
		return "hw.disknames=sd0:e072adf1dcc1be16,cd0:,sd1:4a9f12a79235b9bd\n", nil
	}

	tests := []struct {
		in  string
		out string
		err bool
	}{
		{
			"e072adf1dcc1be16",
			"sd0",
			false,
		},
		{
			"4a9f12a79235b9bd",
			"sd1",
			false,
		},
		{
			"xxxxxxxxxxxxxxxx",
			"", // dysk nie isnieje
			false,
		},
	}

	for i, test := range tests {
		d, err := diskName(test.in)
		if d != test.out {
			t.Errorf("#%d: diskName(): oczekiwano: %q, jest: %q", i, test.out, d)
		}
		e := err != nil
		if e != test.err {
			t.Errorf("#%d: diskName() isErr: oczekiwano: %v, jest: %v", i, test.err, e)
		}
	}
}
