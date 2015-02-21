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
	fs          = flag.String("fs", "", "filesystem")
	dest        = flag.String("dest", "", "katalog docelowy")
	exclude     = flag.String("exclude", "", "co pominąć")
	logfile     = flag.String("logfile", "", "logfile")
	rsync       = flag.String("rsync", "rsync", "polecenie rsync")
	rsync_flags = flag.String("rsync_flags", "-avxH", "flagi polecenia rsync")
)

const (
	dlast     = "last"     // nazwa symlinku do poprzedniego snapshotu
	dsnapshot = "snapshot" // katalog z aktualnym snapshotem
)

const usageStr = `usage: snapshot -fs=filesystem -dest=dir [flags]
	-fs="": kopiowany filesystem
	-dest="": katalog docelowy
	-exclude="": lista wzorców ignorowanych plików (pattern,pattern,...)
	-logfile="": plik z logami
	-rsync="rsync": polecenie 'rsync'
	-rsync_flags="-avxH": flagi polecenia rsync`

func usage() {
	fmt.Fprintln(os.Stderr, usageStr)
	os.Exit(1)
}

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
	flag.Usage = usage
	flag.Parse()

	if err := validateDir(*fs); err != nil {
		fmt.Fprintf(os.Stderr, "zła wartość opcji -fs: %q: %s\n", *fs, err)
		usage()
	}

	if err := validateDir(*dest); err != nil {
		fmt.Fprintf(os.Stderr, "zła wartość opcji -dest: %q: %s\n", *dest, err)
		usage()
	}

	err := snapshot()
	if err != nil {
		os.Exit(2)
	}
}
