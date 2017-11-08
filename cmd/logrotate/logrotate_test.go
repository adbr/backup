// 2015-02-11 Adam Bryt

package main

import (
	"testing"
)

func TestParseLogFile(t *testing.T) {
	tests := []struct {
		file  string
		lf    logFile
		iserr bool
	}{
		{
			"file.log.3",
			logFile{
				name: "file.log",
				num:  3,
				cext: "",
			},
			false,
		},
		{
			"file.log.3.gz",
			logFile{
				name: "file.log",
				num:  3,
				cext: ".gz",
			},
			false,
		},
		{
			"file.abc.log.3.gz",
			logFile{
				name: "file.abc.log",
				num:  3,
				cext: ".gz",
			},
			false,
		},
		{
			"file.log.abc",
			logFile{},
			true, // błąd parsowania "abc"
		},
		{
			"file.log.abc.gz",
			logFile{cext: ".gz"},
			true, // błąd parsowania "abc"
		},
		{
			"a/b/c/file.log.123.gz",
			logFile{
				name: "a/b/c/file.log",
				num:  123,
				cext: ".gz",
			},
			false,
		},
		{
			"file.log.12a.gz",
			logFile{
				cext: ".gz",
			},
			true, // błąd parsowania "12a"
		},
		{
			"file",
			logFile{},
			true, // brak rozszerzenia z numerem
		},
		{
			"file.gz",
			logFile{cext: ".gz"},
			true, // brak rozszerzenia z numerem
		},
	}

	for i, test := range tests {
		lf, err := parseLogFile(test.file)
		iserr := err != nil
		if iserr != test.iserr {
			t.Errorf("#%d: (iserr) oczekiwano: %v, jest: %v: err: %v",
				i, test.iserr, iserr, err)
		}
		if lf != test.lf {
			t.Errorf("#%d: (lf) oczekiwano: %v, jest: %v", i, test.lf, lf)
		}
	}
}
