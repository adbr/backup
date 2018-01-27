// 2015-02-11 Adam Bryt

package logrotate

import (
	"errors"
	"testing"
)

func TestParseLogFile(t *testing.T) {
	// Pomocnicza wartość typu error do zaznaczenia w testach przypadku
	// gdy funkcja powinna zwrócić błąd różny od nil.
	var errParse = errors.New("błąd parsowania nazwy pliku")
	
	var tests = []struct {
		file string   // wejściowa nazwa pliku
		lf   *logFile // wyjściowe składowe nazwy pliku
		err  error    // !nil jeśli powinien wystąpić błąd
	}{
		{
			"file.log.3",
			&logFile{
				name: "file.log",
				num:  3,
				ext:  "",
			},
			nil,
		},
		{
			"file.log.0",
			&logFile{
				name: "file.log",
				num:  0,
				ext:  "",
			},
			nil,
		},
		{
			"file.log.3.gz",
			&logFile{
				name: "file.log",
				num:  3,
				ext:  ".gz",
			},
			nil,
		},
		{
			"file.abc.log.3.gz",
			&logFile{
				name: "file.abc.log",
				num:  3,
				ext:  ".gz",
			},
			nil,
		},
		{
			"a/b/c/file.log.123.gz",
			&logFile{
				name: "a/b/c/file.log",
				num:  123,
				ext:  ".gz",
			},
			nil,
		},
		{
			".5.gz",
			&logFile{
				name: "",
				num:  5,
				ext:  ".gz",
			},
			nil,
		},

		// Przypadki błędne - funkcja powinna zwrócić błąd:

		{
			"file.log", // brak numeru pliku
			nil,
			errParse,
		},
		{
			"file.log.abc", // brak numeru pliku
			nil,
			errParse,
		},
		{
			"file.log.abc.gz", // brak numeru pliku
			nil,
			errParse,
		},
		{
			"file.log.12a.gz", // błędny numer pliku
			nil,
			errParse,
		},
		{
			"file", // brak numeru pliku
			nil,
			errParse,
		},
		{
			"file.gz", // brak numeru pliku
			nil,
			errParse,
		},
	}

	for _, test := range tests {
		lf, err := parseLogFile(test.file)

		// wynik różny od oczekiwanego
		if err == nil && *lf != *test.lf {
			t.Errorf("parseLogFile(%q) = %#v, oczekiwane %#v", test.file, *lf, *test.lf)
		}

		// nie wystąpił oczekiwany błąd
		if err == nil && test.err != nil {
			t.Errorf("parseLogFile(%q) - nie wystąpił oczekiwany błąd", test.file)
		}

		// wystąpił nie oczekiwany błąd
		if err != nil && test.err == nil {
			t.Errorf("parseLogFile(%q) - wystąpił nie oczekiwany błąd: %q", test.file, err)
		}
	}
}
