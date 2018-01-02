// 2015-02-04 Adam Bryt

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/adbr/backup/snapshot"
)

func main() {
	src := flag.String("src", "", "")
	dst := flag.String("dst", "", "")
	exclude := flag.String("exclude", "", "")
	logfile := flag.String("logfile", "", "")
	rsync := flag.String("rsync", "rsync", "")
	rsyncopts := flag.String("rsyncopts", "-avxH8", "")
	h := flag.Bool("h", false, "")
	help := flag.Bool("help", false, "")

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

	if *src == "" {
		fmt.Fprintln(os.Stderr, "snapshot: brakuje opcji -src")
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}
	if *dst == "" {
		fmt.Fprintln(os.Stderr, "snapshot: brakuje opcji -dst")
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}

	if *logfile != "" {
		file, err := os.OpenFile(*logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "snapshot: logfile: %s\n", err)
			os.Exit(2)
		}
		snapshot.LogFile = file
	}

	snapshot.RsyncCommand = *rsync
	snapshot.RsyncOptions = *rsyncopts
	err := snapshot.Snapshot(*src, *dst, *exclude)
	if err != nil {
		fmt.Fprintf(os.Stderr, "snapshot: %s\n", err)
		os.Exit(1)
	}
}

// Stała usageText zawiera opis opcji programu wyświetlany przy użyciu
// opcji -h lub w przypadku błędu parsowania opcji.
const usageText = `Sposób użycia:
	snapshot [opcje] -src=filesystem -dst=directory
Opcje:
	-src filesystem
		backupowany filesystem
	-dst directory
		docelowy katalog z backupami
	-exclude string
		lista wzorców ignorowanych plików "pattern,pattern,..."
		(domyślnie: "")
	-logfile filename
		plik z logami (domyślnie: "")
	-rsync filename
		nazwa polecenia rsync (domyślnie: "rsync")
	-rsyncopts string
		opcje polecenia rsync (domyślnie: "-avxH8")
	-h	sposób użycia
	-help	dokumentacja
`

// Stała helpText zawiera opis programu wyświetlany przy użyciu opcji
// -help. Treść jest identyczna jak w doc comment programu z pliku
// doc.go.
const helpText = `
Program snapshot kopiuje katalog przy użyciu polecenia rsync(1). Tworzy
kolejne snapshoty w katalogach o nazwach typu '2015-02-10T18:07:39'.
Jeśli pliki w aktualnym snapshocie nie zmieniły się, to nie są
kopiowane tylko są tworzone hard linki do plików w poprzednim
snapshocie.

Sposób użycia:
	snapshot [opcje] -src=filesystem -dst=directory
Opcje:
	-src filesystem
		backupowany filesystem
	-dst directory
		docelowy katalog z backupami
	-exclude string
		lista wzorców ignorowanych plików "pattern,pattern,..."
		(domyślnie: "")
	-logfile filename
		plik z logami (domyślnie: "")
	-rsync filename
		nazwa polecenia rsync (domyślnie: "rsync")
	-rsyncopts string
		opcje polecenia rsync (domyślnie: "-avxH8")
	-h	sposób użycia
	-help	dokumentacja

Do kopiowania jest używane polecenie rsync(1) z następującymi opcjami:

	-a			archive mode
	-v			verbose
	-x			don't cross filesystem boundaries
	-H			preserve hard links
	-8			8-bit output
	--link-dest=DIR		hardlink to files in DIR when unchanged
	--exclude=PATTERN	exclude files matching PATTERN
`
