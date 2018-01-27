// 2017-11-11 adbr

// Pakiet snapshot grupuje funkcje do tworzenia snaphotu filesystemu
// przy użyciu polecenia rsync.
package snapshot

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	RsyncCommand = "rsync"
	RsyncOptions = "-avxH8"
	LogFile      *os.File
)

// Snapshot kopiuje katalog src do dst używając polecenia rsync(1);
// pomija pliki pasujące do wzorców w exclude. Argument exclude
// zawiera listę wzorców ignorowanych plików w postaci
// "pattern,pattern,...".
func Snapshot(src, dst, exclude string) error {
	info("=== początek snapshotu (%s)", timestamp())
	info("src: %q", src)
	info("dst: %q", dst)
	begin := time.Now()

	// utworzenie tymczasowego katalogu snapshot
	snapshotdir, err := makeSnapshotDir(dst)
	if err != nil {
		return err
	}

	// przygotowanie argumentów polecenia rsync
	var args []string

	// opcje standardowe
	args = append(args, strings.Fields(RsyncOptions)...)

	// opcje exclude
	opts := excludeOptions(exclude)
	if len(opts) != 0 {
		args = append(args, opts...)
	}

	// opcja linkdest
	opt, err := linkdestOption(dst)
	if err != nil {
		return err
	}
	if len(opt) != 0 {
		args = append(args, opt)
	}

	// argumenty katalogi
	args = append(args, src, snapshotdir)

	// uruchomienie polecenia rsync
	cmd := exec.Command(RsyncCommand, args...)
	info("polecenie: %q", strings.Join(cmd.Args, " "))
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	// czytanie i logowanie wyjścia z rsync
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		info("rsync: %s", line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	// zmiana nazwy katalogu ze snapshotem na timestamp
	timestamp := timestamp()
	timestampdir := filepath.Join(dst, timestamp)
	info("zmiana nazwy katalogu %q na %q", "snapshot", timestamp)
	err = os.Rename(snapshotdir, timestampdir)
	if err != nil {
		return err
	}

	// ustawienie symlinku 'last' na ostatni snapshot
	info("zmiana symlinku %q -> %q", "last", timestamp)
	lastdir := filepath.Join(dst, "last")
	err = os.Remove(lastdir)
	if err != nil {
		if os.IsNotExist(err) {
			info("warning: katalog %q nie istnieje - pierwszy snapshot?", lastdir)
		} else {
			return err
		}
	}
	err = os.Symlink(timestamp, lastdir)
	if err != nil {
		return err
	}

	end := time.Now()
	info("koniec snapshotu, czas trwania: %s", end.Sub(begin))
	return nil
}

// excludeOptions parsuje string patterns zawierający listę wzorców i
// zwraca listę opcji --exclude dla rsync. Argument patterns jest
// stringiem zawierającym wzorce oddzielone przecinkami np.:
// "adbr/tmp/*,.cache/*".
func excludeOptions(patterns string) []string {
	var opts []string
	if patterns == "" {
		return opts
	}
	a := strings.Split(patterns, ",")
	for _, pat := range a {
		opts = append(opts, "--exclude="+pat)
	}
	return opts
}

// linkdestOption tworzy i zwraca opcję '--link-dest' dla programu
// rsync. jeśli w katalogu dst istnieje symlink 'last' wskazujący na
// katalog z poprzednim snapshotem to zwraca utworzoną opcję. Jeśli
// 'last' nie istnieje to loguje komunikat i zwraca string pusty.
// Argument dst jest katalogiem docelowym, czyli katalogiem w którym
// tworzone są snapshoty.
func linkdestOption(dst string) (string, error) {
	lastdir := filepath.Join(dst, "last")
	// opcja --link-dest wymaga żeby jej argument był bezwzględną
	// nazwą katalogu - jeśli nie jest to nie widzi katalogu
	lastdir, err := filepath.Abs(lastdir)
	if err != nil {
		return "", err
	}

	fi, err := os.Stat(lastdir)
	if err != nil {
		if os.IsNotExist(err) {
			info("warning: katalog %q nie istnieje - pierwszy snapshot?", lastdir)
			return "", nil
		}
		return "", err
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("%q nie jest katalogiem", lastdir)
	}

	return "--link-dest=" + lastdir, nil
}

// info loguje sformatowany komunikat do stdout i pliku LogFile jeśli
// LogFile jest różny od nil. W przypadku błędu wywołuje panic.
func info(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	format = "snapshot: " + format
	fmt.Printf(format, args...)
	if LogFile != nil {
		_, err := fmt.Fprintf(LogFile, format, args...)
		if err != nil {
			panic(err)
		}
	}
}

// makeSnapshotDir tworzy w docelowym katalogu dst, tymczasowy katalog
// roboczy o nazwie 'snapshot', w którym będzie wykonywany aktualny
// snapshot. Zwraca bezwzględną nazwę utworzonego katalogu i błąd
// jeśli wystąpił.
func makeSnapshotDir(dst string) (string, error) {
	dir := filepath.Join(dst, "snapshot")
	err := os.Mkdir(dir, 0755)
	if err != nil {
		if os.IsExist(err) {
			info("warning: katalog \"snapshot\" już istnieje - nie dokończony poprzedni snapshot?")
			return dir, nil
		}
		return "", err
	}
	return dir, nil
}

// timestamp zwraca string z aktualną datą i czasem w formacie
// 'yyyy-mm-ddThh:mm:ss'.
func timestamp() string {
	t := time.Now()
	return t.Format("2006-01-02T15:04:05")
}
