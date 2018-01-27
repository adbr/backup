// 2018-01-05 adbr

// Pakiet logrotate umożliwia archiwizowanie plików z logami po
// przekroczeniu zadanego rozmiaru pliku. Przechowywana jest określona
// liczba poprzednio zarchiwizowanych plików. Archiwizacja polega na
// zmianie nazwy pliku, np: file.log -> file.log.0, file.log.0 ->
// file.log.1, itd.
package logrotate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Zmienna Verbose włącza logowanie na stdout komunikatów o
// wykonywanych czynnościach.
var Verbose = false

// Typ logFile reprezentuje składowe nazwy pliku z logami. Na przykład
// logFile{name: "filename.log", num: 1, ext: ".gz"} odpowiada nazwie
// pliku "filename.log.1.gz".
type logFile struct {
	name string // główna nazwa pliku, bez numeru i ext
	num  int    // numer pliku
	ext  string // rozszerzenie dla pliku skompresowanego (np. ".gz")
}

// parseLogFile parsuje nazwę pliku i zwraca logFile z poszczególnymi
// częściami składowymi nazwy. Na przykład dla pliku
// "filename.log.1.gz" zwróci logFile{name: "filename.log", num: 1,
// ext: ".gz"}. Rozszeszenie z numerem pliku musi wystąpić. Opcjonalne
// rozszerzenie pliku skompresowanego może być tylko ".gz".
func parseLogFile(file string) (*logFile, error) {
	var logfile logFile

	// rozszerzenie .gz (opcjonalne)
	ext := filepath.Ext(file)
	if ext == ".gz" {
		logfile.ext = ext
		file = file[:len(file)-len(ext)]
	}

	// rozszerzenie z numerem (wymagane)
	ext = filepath.Ext(file)
	s := strings.TrimLeft(ext, ".")
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("parse log fname: %s", err)
	}
	logfile.num = n
	file = file[:len(file)-len(ext)]

	// główna nazwa pliku
	logfile.name = file

	return &logfile, nil
}

// Rotate sprawdza czy plik file ma rozmiar większy od size, i jeśli
// tak to robi rotację plików z logami zachowując num najnowszych
// plików. Jeśli size = 0 to plik nie jest rotowany; jeśli num = 0 to
// nie ma ograniczenia na liczbę archiwizowanych plików.
func Rotate(file string, size int64, num int) error {
	// sprawdzenie czy plik jest gotowy do archiwizacji
	ok, err := isReady(file, size)
	if err != nil {
		return err
	}
	if !ok {
		info("plik %q nie jest gotowy do rotacji", file)
		return nil
	}

	// wczytanie nazw plików zarchiwizowanych
	a, err := globLogFiles(file)
	if err != nil {
		return err
	}

	// sortowanie plików według numerów
	sort.Slice(a, func(i, j int) bool {
		return a[i].num < a[j].num
	})

	// rotacja plików - zaczynając od ostatniego
	for i := len(a) - 1; i >= 0; i-- {
		if num > 0 && a[i].num >= num-1 {
			continue
		}
		err := renameLogFile(a[i])
		if err != nil {
			return err
		}
	}

	// zmiana nazwy głównego pliku z logami
	new := file + ".0"
	info("%q -> %q", file, new)
	err = os.Rename(file, new)
	if err != nil {
		return err
	}

	// utworzenie pustego pliku - truncate file
	info("utworzenie pustego pliku %q", file)
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	f.Close()

	return nil
}

// isReady sprawdza czy plik z logami jest gotowy do archiwizacji.
func isReady(file string, size int64) (bool, error) {
	info, err := os.Stat(file)
	if err != nil {
		return false, err
	}
	if size <= 0 {
		return false, nil
	}
	if info.Size() >= size {
		return true, nil
	}
	return false, nil
}

// renameLogFile zmienia nazwę pliku z logami file na nazwę z numerem
// o 1 większym.
func renameLogFile(file *logFile) error {
	old := fmt.Sprintf("%s.%d%s", file.name, file.num, file.ext)
	new := fmt.Sprintf("%s.%d%s", file.name, file.num+1, file.ext)
	info("%q -> %q", old, new)
	err := os.Rename(old, new)
	if err != nil {
		return err
	}
	return nil
}

// globLogFile zwraca slice z plikami pasującymi do wzorca "file.*" i
// spełniającymi warunki parseLogFile, czyli archiwami pliku z logami
// file.
func globLogFiles(file string) ([]*logFile, error) {
	var files []*logFile

	names, err := filepath.Glob(file + ".*")
	if err != nil {
		return files, err
	}

	for _, s := range names {
		f, err := parseLogFile(s)
		if err != nil {
			return files, err
		}
		files = append(files, f)
	}
	return files, nil
}

// info drukuje sformatowany komunikat do stdout. Dodaje prefiks i
// znak nowego wiersza.
func info(format string, args ...interface{}) {
	if !Verbose {
		return
	}
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	format = "logrotate: " + format
	fmt.Printf(format, args...)
}
