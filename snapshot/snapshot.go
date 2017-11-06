// 2015-02-04 Adam Bryt

package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

var (
	fs          = flag.String("fs", "", "")
	dest        = flag.String("dest", "", "")
	exclude     = flag.String("exclude", "", "")
	logfile     = flag.String("logfile", "", "")
	rsync       = flag.String("rsync", "rsync", "")
	rsync_flags = flag.String("rsync_flags", "-avxH8", "")
	h           = flag.Bool("h", false, "")
	help        = flag.Bool("help", false, "")
)

const (
	dlast     = "last"     // nazwa symlinku do poprzedniego snapshotu
	dsnapshot = "snapshot" // katalog z aktualnym snapshotem
)

func parseExclude(s string) []string {
	var opts []string
	a := strings.Split(s, ",")
	for _, pat := range a {
		opts = append(opts, "--exclude="+pat)
	}
	return opts
}

func timestamp() string {
	t := time.Now()
	return t.Format("2006-01-02T15:04:05")
}

func logBuf(prefix string, buf *bytes.Buffer) {
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("%s%s", prefix, line)
	}
}

func rsyncSnapshot() error {
	var cmdargs = []string{*rsync}

	rsyncFlags := strings.Fields(*rsync_flags)
	cmdargs = append(cmdargs, rsyncFlags...)

	// utworzenie katalogu snapshot
	snapdir := path.Join(*dest, dsnapshot)
	log.Printf("utworzenie katalogu '%s'", dsnapshot)
	err := os.Mkdir(snapdir, os.ModePerm)
	if err != nil {
		if os.IsExist(err) {
			msg := "WARNING: katalog '%s' już istnieje: " +
				"nie dokończony poprzedni snapshot?"
			log.Printf(msg, snapdir)
		} else {
			return err
		}
	}

	// utworzenie opcji --link-dest
	lastdir := path.Join(*dest, dlast)
	info, err := os.Stat(lastdir)
	if err != nil {
		if os.IsNotExist(err) {
			msg := "WARNING: katalog '%s' nie istnieje: pierwszy snapshot?"
			log.Printf(msg, lastdir)
		} else {
			return err
		}
	} else {
		if !info.IsDir() {
			return fmt.Errorf("'%s' nie jest katalogiem", lastdir)
		}
		opt := "--link-dest=" + lastdir
		cmdargs = append(cmdargs, opt)
	}

	// utworzenie opcji --exclude
	if *exclude != "" {
		a := parseExclude(*exclude)
		cmdargs = append(cmdargs, a...)
	}

	// dodanie argumentów polecenia
	cmdargs = append(cmdargs, *fs, snapdir)

	// uruchomienie rsync

	cmd := exec.Command(cmdargs[0], cmdargs[1:]...)
	var errbuf = new(bytes.Buffer)
	cmd.Stderr = errbuf
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	cmdstr := strings.Join(cmd.Args, " ")
	log.Printf("kopiowanie danych poleceniem: '%s'", cmdstr)

	err = cmd.Start()
	if err != nil {
		logBuf("rsync stderr: ", errbuf)
		return fmt.Errorf("błąd uruchomienia polecenia '%s': %s", cmdstr, err)
	}

	// drukowanie wyjścia z rsync
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("rsync: %s", line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		logBuf("rsync stderr: ", errbuf)
		return fmt.Errorf("błąd polecenia '%s': %s", cmdstr, err)
	}

	// zmiana nazwy katalogu ze snapshotem
	ts := timestamp()
	tsdir := path.Join(*dest, ts)
	log.Printf("zmiana nazwy katalogu '%s' na '%s'", dsnapshot, ts)
	err = os.Rename(snapdir, tsdir)
	if err != nil {
		return err
	}

	// ustawienie last na nowy snapshot
	log.Printf("zmiana symlinku '%s' -> '%s'", dlast, ts)
	err = os.Remove(lastdir)
	if err != nil {
		if os.IsNotExist(err) {
			msg := "WARNING: katalog '%s' nie istnieje: pierwszy snapshot?"
			log.Printf(msg, lastdir)
		} else {
			return err
		}
	}
	err = os.Symlink(ts, lastdir)
	if err != nil {
		return err
	}

	return nil
}

func snapshot() error {
	log.SetPrefix("snapshot: ")
	if *logfile != "" {
		f, err := os.OpenFile(*logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Printf("ERROR: %s", err)
			return err
		}
		defer f.Close()
		w := io.MultiWriter(os.Stderr, f)
		log.SetOutput(w)
	}

	log.Print("początek snapshotu")
	log.Printf("filesystem: '%s'", *fs)
	log.Printf("katalog docelowy: '%s'", *dest)
	begin := time.Now() // czas początku snapshotu

	err := rsyncSnapshot()
	if err != nil {
		log.Printf("ERROR: %s", err)
		return err
	}

	log.Print("koniec snapshotu")
	d := time.Since(begin)
	log.Printf("czas trwania snapshotu: %s", d)

	return nil
}

func validateDir(dir string) error {
	if dir == "" {
		return errors.New("nazwa katalogu jest pusta")
	}

	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%q nie jest katalogiem", dir)
	}

	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usageText)
	}
	flag.Parse()

	if *h {
		fmt.Print(usageText)
		return
	}
	if *help {
		fmt.Print(helpText)
		return
	}

	// sprawdzanie wymaganych opcji
	if err := validateDir(*fs); err != nil {
		fmt.Fprintf(os.Stderr, "zła wartość opcji -fs: %s\n", err)
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}
	if err := validateDir(*dest); err != nil {
		fmt.Fprintf(os.Stderr, "zła wartość opcji -dest: %s\n", err)
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}

	err := snapshot()
	if err != nil {
		os.Exit(1)
	}
}

// Stała usageText zawiera opis opcji programu wyświetlany przy użyciu
// opcji -h lub w przypadku błędu parsowania opcji.
const usageText = `Sposób użycia:
	snapshot [opcje] -fs=filesystem -dest=directory
Opcje:
	-fs filesystem
		backupowany filesystem
	-dest directory
		katalog docelowy
	-exclude string
		lista wzorców ignorowanych plików "pattern,pattern,..."
		(domyślnie: "")
	-logfile filename
		plik z logami (domyślnie: "")
	-rsync filename
		nazwa polecenia rsync (domyślnie: "rsync")
	-rsync_flags string
		opcje polecenia rsync (domyślnie: "-avxH8")
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
	snapshot [opcje] -fs=filesystem -dest=directory
Opcje:
	-fs filesystem
		backupowany filesystem
	-dest directory
		katalog docelowy
	-exclude string
		lista wzorców ignorowanych plików "pattern,pattern,..."
		(domyślnie: "")
	-logfile filename
		plik z logami (domyślnie: "")
	-rsync filename
		nazwa polecenia rsync (domyślnie: "rsync")
	-rsync_flags string
		opcje polecenia rsync (domyślnie: "-avxH8")

Do kopiowania jest używane polecenie rsync(1) z następującymi opcjami:

	-a			archive mode
	-v			verbose
	-x			don't cross filesystem boundaries
	-H			preserve hard links
	-8			8-bit output
	--link-dest=DIR		hardlink to files in DIR when unchanged
	--exclude=PATTERN	exclude files matching PATTERN
`
