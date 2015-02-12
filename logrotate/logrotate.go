// 2015-02-11 Adam Bryt

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var (
	num  = flag.Int("num", 0, "liczba archiwizowanych plików (0 bez ograniczeń)")
	size = flag.Int64("size", 0, "wielkość (w bajtach) archiwizowanego pliku (0 bez rotacji)")
)

const usageStr = "usage: logrotate [flags] logfile"

func usage() {
	fmt.Fprintln(os.Stderr, usageStr)
	flag.PrintDefaults()
	os.Exit(1)
}

// Rozszerzenia skompresowanych archiwalnych plików z logami.
var cexts = []string{".gz"}

// isCext sprawdza czy rozszerzenie ext jest dozwolonym rozszerzeniem
// pliku skompresowanego.
func isCext(ext string) bool {
	for _, e := range cexts {
		if ext == e {
			return true
		}
	}
	return false
}

// Typ logFile opisuje zarchiwizowany plik z logami.
type logFile struct {
	name string // nazwa pliku z logami bez numeru i cext
	num  int    // numer pliku
	cext string // rozszerzenie pliku skompresowanego (np. ".gz")
}

// parseLogFile parsuje nazwę zarchiwizowanego pliku z logami.
// Np. dla pliku "log.3.gz": name == "log", num == 3, cext == ".gz"
func parseLogFile(file string) (logFile, error) {
	f := file
	var lf logFile

	ext := path.Ext(f)

	// może zawierać opcjonalne roszerzenie .gz
	if isCext(ext) {
		lf.cext = ext
		f = strings.TrimSuffix(f, ext)
		ext = path.Ext(f)
	}

	if ext == "" {
		return lf, fmt.Errorf("parseLogFile: brak rozszerzenia z numerem: %q", file)
	}

	s := strings.TrimLeft(ext, ".")
	num, err := strconv.Atoi(s)
	if err != nil {
		return lf, fmt.Errorf("parseLogFile: %s: %s", file, err)
	}
	lf.num = num

	f = strings.TrimSuffix(f, ext)
	lf.name = f

	return lf, nil
}

// Typ logFileSlice zawiera listę opisującą zarchiwizowane pliki z logami.
// Dla sortowania plików z logami po numerze pliku.
type logFileSlice []logFile

// Implementacja sort.Interface
func (p logFileSlice) Len() int {
	return len(p)
}

// Implementacja sort.Interface
func (p logFileSlice) Less(i, j int) bool {
	return p[i].num < p[j].num
}

// Implementacja sort.Interface
func (p logFileSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// globLogFile zwraca slice z plikami pasującymi do wzorca 'file.*' i
// spełniającymi warunki parseLogFile.
func globLogFile(file string) (logFileSlice, error) {
	var lfs logFileSlice

	a, err := filepath.Glob(file + ".*")
	if err != nil {
		return lfs, err
	}

	for _, f := range a {
		lf, err := parseLogFile(f)
		if err != nil {
			return lfs, err
		}
		lfs = append(lfs, lf)
	}

	return lfs, nil
}

// rotateFile zmienia nazwę pliku z logami na nazwę z kolejnym numerem.
func rotateFile(file logFile) error {
	old := fmt.Sprintf("%s.%d%s", file.name, file.num, file.cext)
	new := fmt.Sprintf("%s.%d%s", file.name, file.num+1, file.cext)
	err := os.Rename(old, new)
	if err != nil {
		return err
	}
	return nil
}

// isReady sprawdza czy plik z logami jest gotowy do archiwizacji.
func isReady(file string) (bool, error) {
	if *size <= 0 {
		return false, nil
	}

	fi, err := os.Stat(file)
	if err != nil {
		return false, err
	}

	if fi.Size() >= *size {
		return true, nil
	}

	return false, nil
}

func logrotate(file string) error {
	// wczytanie nazw plików zarchiwizowanych
	a, err := globLogFile(file)
	if err != nil {
		return err
	}

	// posortowanie zarchiwizowanych plików według numerów
	sort.Sort(a)

	// rotacja plików - zaczynając od ostatniego
	for i := len(a) - 1; i >= 0; i-- {
		if *num > 0 && i >= *num-1 {
			continue
		}
		err := rotateFile(a[i])
		if err != nil {
			return err
		}
	}

	// zmiana nazwy głównego pliku z logami
	new := file + ".0"
	err = os.Rename(file, new)
	if err != nil {
		return err
	}

	// utworzenie pustego pliku
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	f.Close()

	return nil
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
	}

	file := flag.Arg(0)

	// sprawdzenie czy plik istnieje
	_, err := os.Stat(file)
	if err != nil {
		log.Fatal(err)
	}

	ok, err := isReady(file)
	if err != nil {
		log.Fatal(err)
	}
	if ok {
		err := logrotate(file)
		if err != nil {
			log.Fatal(err)
		}
	}
}
