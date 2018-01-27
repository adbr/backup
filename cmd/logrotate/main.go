// 2015-02-11 Adam Bryt

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/adbr/backup/internal/logrotate"
)

func main() {
	num := flag.Int("num", 0, "")
	size := flag.Int64("size", 0, "")
	v := flag.Bool("v", false, "verbose")
	h := flag.Bool("h", false, "usage")
	help := flag.Bool("help", false, "help")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usageText)
	}
	flag.Parse()

	if *h {
		fmt.Print(usageText)
		os.Exit(0)
	}
	if *help {
		fmt.Print(helpText)
		os.Exit(0)
	}
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "logrotate: brak argumentu logfile")
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}

	file := flag.Arg(0)
	logrotate.Verbose = *v
	err := logrotate.Rotate(file, *size, *num)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logrotate: %s\n", err)
		os.Exit(1)
	}
}

// Stała usageText zawiera opis opcji programu wyświetlany przy użyciu
// opcji -h lub w przypadku błędu parsowania opcji.
const usageText = `Sposób użycia:
	logrotate [opcje] logfile
Opcje:
	-num int
		maksymalna liczba archiwizowanych plików (domyślnie:
		0, czyli bez ograniczenia)
	-size int
		wielkość (w bajtach) archiwizowanego pliku, jeśli
		rozmiar pliku logfile jest większy niż -size to plik
		jest archiwizowany (domyślnie: 0, czyli bez
		ograniczenia)
	-v	wyświetlanie komunikatów (verbose)
	-h	sposób użycia
	-help	dokumentacja
`

// Stała helpText zawiera opis programu wyświetlany przy użyciu opcji
// -help. Treść jest identyczna jak w doc comment programu z pliku
// doc.go.
const helpText = `
Program logrotate służy do rotacji plików z logami. Jeśli rozmiar
pliku jest większy niż wartość podana w opcji -size to plik jest
archiwizowany: nazwa pliku log jest zmieniana na log.0 i tworzony jest
pusty plik log.

Jeśli istnieją wcześniej zarchiwizowane pliki z logami to ich numery
są zwiększane, np:

	log.0 -> log.1
	log.1 -> log.2
	...

Maksymalną liczbę zarchiwizowanych plików z logami określa opcja -num.

Sposób użycia:
	logrotate [opcje] logfile
Opcje:
	-num int
		maksymalna liczba archiwizowanych plików (domyślnie:
		0, czyli bez ograniczenia)
	-size int
		wielkość (w bajtach) archiwizowanego pliku, jeśli
		rozmiar pliku logfile jest większy niż -size to plik
		jest archiwizowany (domyślnie: 0, czyli bez
		ograniczenia)
	-v	wyświetlanie komunikatów (verbose)
	-h	sposób użycia
	-help	dokumentacja
`
