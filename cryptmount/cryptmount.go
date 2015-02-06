// 2015-01-27 Adam Bryt

// Program cryptmount podłącza szyfrowaną partycję do softraid a
// następnie montuje partycję na szyfrowanym dysku do filesystemu.
//
// Jest przeznaczony dla systemu OpenBSD. Do obsługi szyfrowanej
// partycji jest używany softraid(4) i polecenie bioctl(8).
//
// Celem programu jest ułatwienie montowania dysku przez DUID -
// polecenie bioctl musi mieć nazwę dysku z /dev. Program tłumaczy
// nazwy DUID na aktualne nazwy dysków w katalogu /dev używając
// polecenia 'sysctl hw.disknames'.
//
// Sposób użycia:
//
//	cryptmount [flags] -disk0=diskspec -disk1=diskspec dir
//
// Wartości opcji disk0 i disk1 (diskspec) mają format 'DUID.PART',
// gdzie DUID jest unikalnym identyfikatorem dysku z disklabel, a PART
// jest pojedynczą literą oznaczającą partycję na tym dysku (np.
// "a3a6acb427840bc0.a").
//
// Opcja -disk0 specyfikuje zaszyfrowaną partycję typu RAID na dysku
// fizycznym (np. USB), która zostanie podłączona do softraid0.
// Opcja -disk1 specyfikuje partycję typu FFS na rozszyfrowanym dysku
// logicznym, która zostanie podmontowana w katologu dir.
//
// Opcje:
//
//	-disk0=""
//		dysk i partycja zaszyfrowana typu RAID do podłączenie
//		do softraid (opcja wymagana)
//	-disk1=""
//		dysk i partycja typu FFS na rozszyfrowanym dysku logicznym
//		do podmontowania w katalogu dir (opcja wymagana)
//	-mountopts="softdep"
//		opcje polecenia mount
//	-u=false
//		odmontuj dyski (unmount)
//	-h=false
//		wyświetl help
//
package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	disk0   = flag.String("disk0", "", "dysk i partycja zaszyfrowana typu RAID")
	disk1   = flag.String("disk1", "", "dysk i partycja typu FFS")
	logfile = flag.String("log", "", "logfile")
	mntopts = flag.String("mountopts", "softdep", "opcje polecenia mount")
	uFlag   = flag.Bool("u", false, "odmontuj dyski (unmount)")
	hFlag   = flag.Bool("h", false, "wyświetl help")
)

const usageStr = `usage: cryptmount [flags] -disk0=diskspec -disk1=diskspec dir
	-disk0=""
		dysk i partycja zaszyfrowana typu RAID do podłączenie
		do softraid (opcja wymagana)
	-disk1=""
		dysk i partycja typu FFS na rozszyfrowanym dysku logicznym
		do podmontowania w katalogu dir (opcja wymagana)
	-mountopts="softdep"
		opcje polecenia mount
	-u=false
		odmontuj dyski (unmount)
	-h=false
		wyświetl help`

const helpStr = `Wartości opcji disk0 i disk1 (diskspec) mają format 'DUID.PART',
gdzie DUID jest unikalnym identyfikatorem dysku z disklabel, a PART
jest pojedynczą literą oznaczającą partycję na tym dysku (np.
"a3a6acb427840bc0.a").

Opcja -disk0 specyfikuje zaszyfrowaną partycję typu RAID na dysku
fizycznym (np. USB), która zostanie podłączona do softraid0.
Opcja -disk1 specyfikuje partycję typu FFS na rozszyfrowanym dysku
logicznym, która zostanie podmontowana w katologu dir.`

func usage() {
	fmt.Fprintln(os.Stderr, usageStr)
	os.Exit(2)
}

func help() {
	fmt.Fprintln(os.Stdout, usageStr)
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, helpStr)
	os.Exit(0)
}

// getDiskNames zwraca wynik polecenia 'sysctl hw.disknames',
// czyli listę dysków w postaci:
// "hw.disknames=sd0:e072adf1dcc1be16,cd0:,sd1:4a9f12a79235b9bd"
// Funkcja jest zmienną dla ułatwienia testów.
var getDiskNames = func() (string, error) {
	// bufor na komunikaty wysyłane do stderr - dla obsługi przypadku gdy
	// exit status jest 0 ale na stderr został zwrócony komunikat o błędzie
	var eout bytes.Buffer

	cmd := exec.Command("sysctl", "hw.disknames")
	cmd.Stderr = &eout
	out, err := cmd.Output()
	if err != nil {
		if eout.Len() != 0 {
			return "", fmt.Errorf("%s: %s", eout.String(), err)
		} else {
			return "", err
		}
	}

	// obsługa przypadku gdy exit status jest 0 ale na stderr został
	// zwrócony komunikat o błędzie
	if eout.Len() != 0 {
		return "", errors.New(eout.String())
	}

	s := string(out)
	return s, nil
}

var errNoDisk = errors.New("dysk nie podłączony")

// diskName zwraca nazwę dysku w /dev odpowiadającą unikalnemu
// identyfikatorowi dysku duid. Jeśli dysk o danym duid nie istnieje to
// zwraca errNoDisk. Np. dla duid 'e072adf1dcc1be16' zwróci 'sd0'.
func diskName(duid string) (string, error) {
	disks, err := getDiskNames()
	if err != nil {
		return "", err
	}
	disks = strings.TrimSpace(disks)

	a := strings.Split(disks, "=")
	if len(a) != 2 {
		return "", fmt.Errorf("nie poprawna lista dysków: %q", disks)
	}

	a = strings.Split(a[1], ",")
	for _, d := range a {
		a := strings.Split(d, ":")
		if len(a) != 2 {
			return "", fmt.Errorf("nie poprawna lista dysków: %q", disks)
		}
		if a[1] == duid {
			return a[0], nil
		}
	}

	return "", errNoDisk
}

func split(diskspec string) (duid string, part string) {
	a := strings.Split(diskspec, ".")
	duid, part = a[0], a[1]
	return
}

// diskNameFull zwraca nazwę pliku specjalnego w /dev odpowiadającą dyskowi
// disk. Parametr disk ma postać duid.part. Np. dla disk 'e072adf1dcc1be16.a'
// zwróci '/dev/sd0a'. Jeśli dysk nie istnieje to zwraca errNoDisk.
func diskNameFull(disk string) (string, error) {
	duid, part := split(disk)

	dev, err := diskName(duid)
	if err != nil {
		return "", err
	}

	name := "/dev/" + dev + part
	return name, nil
}

// Funckcja mountSoftraid podłącza szyfrowany dysk do softraid. Parametr disk
// ma postać duid.part.
func mountSoftraid(disk string) error {
	dev, err := diskNameFull(disk)
	if err == errNoDisk {
		return fmt.Errorf("mountSoftraid: dysk nie podłączony: %q", disk)
	}
	if err != nil {
		return fmt.Errorf("mountSoftraid: %s", err)
	}

	// polecenie: bioctl -c C -l /dev/sd2a softraid0
	cmd := exec.Command("bioctl", "-c", "C", "-l", dev, "softraid0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		c := strings.Join(cmd.Args, " ")
		return fmt.Errorf("mountSoftraid: błąd polecenia: %q: %s", c, err)
	}

	return nil
}

// Funkcja unmountSoftraid odłącza szyfrowany dysk od softraid. Parametr disk
// ma postać duid.part i powinien być dyskiem logicznym (rozszyfrowanym).
func unmountSoftraid(disk string) error {
	duid, _ := split(disk)
	dev, err := diskName(duid)
	if err == errNoDisk {
		return fmt.Errorf("unmountSoftraid: dysk nie podłączony: %q", disk)
	}
	if err != nil {
		return fmt.Errorf("unmountSoftraid: %s", err)
	}

	// polecenie: bioctl -d sd3
	cmd := exec.Command("bioctl", "-d", dev)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		c := strings.Join(cmd.Args, " ")
		return fmt.Errorf("unmountSoftraid: błąd polecenia: %q: %s", c, err)
	}

	return nil
}

// isMountedSoftraid sprawdza czy dysk jest podpięty do softraid. Parametr disk
// ma postać duid.part.
func isMountedSoftraid(disk string) (bool, error) {
	duid, part := split(disk)
	dev, err := diskName(duid)
	if err == errNoDisk {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("isMountedSoftraid: %s", err)
	}

	// polecenie: bioctl softraid0
	cmd := exec.Command("bioctl", "softraid0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		c := strings.Join(cmd.Args, " ")
		out = bytes.TrimSpace(out)
		if len(out) == 0 {
			return false, fmt.Errorf("isMountedSoftraid: błąd polecenia '%s': %s", c, err)
		} else {
			return false, fmt.Errorf("isMountedSoftraid: błąd polecenia '%s': %s: %s", c, err, out)
		}
	}

	d := "<" + dev + part + ">" // nazwa dysku na liście 'bioctl softraid0'
	buf := bytes.NewBuffer(out)
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		line := scanner.Text()
		a := strings.Fields(line)
		if len(a) == 6 && a[5] == d {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("isMountedSoftraid: %s", err)
	}

	return false, nil
}

// mountFFS montuje filesystem na dysku disk w katalogu dir. Argument
// disk a postać duid.part.
func mountFFS(disk string, dir string) error {
	dev, err := diskNameFull(disk)
	if err == errNoDisk {
		return fmt.Errorf("mountFFS: dysk nie podłączony: %s", disk)
	}
	if err != nil {
		return fmt.Errorf("mountFFS: %s", err)
	}

	// polecenie: mount -o softdep /dev/sd2a /backup
	cmd := exec.Command("mount", "-o", *mntopts, dev, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		c := strings.Join(cmd.Args, " ")
		return fmt.Errorf("mountFFS: błąd polecenia: %q: %s", c, err)
	}

	return nil
}

// unmountFFS odmontowuje filesystem na dysku disk. Argument disk ma
// postać duid.part.
func unmountFFS(disk string) error {
	dev, err := diskNameFull(disk)
	if err == errNoDisk {
		return fmt.Errorf("unmountFFS: dysk nie podłączony: %s", disk)
	}
	if err != nil {
		return fmt.Errorf("unmountFFS: %s", err)
	}

	// polecenie: umount /dev/sd2a
	cmd := exec.Command("umount", dev)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		c := strings.Join(cmd.Args, " ")
		return fmt.Errorf("unmountFFS: błąd polecenia: %q: %s", c, err)
	}

	return nil
}

// isMountedFFS sprawdza czy filesystem na dysku logicznym
// (rozszyfrowanym) jest podmontowany w katalogu dir.
func isMountedFFS(disk string, dir string) (bool, error) {
	dev, err := diskNameFull(disk)
	if err == errNoDisk {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("isMountedFFS: %s", err)
	}

	// polecenie: mount
	cmd := exec.Command("mount")
	out, err := cmd.Output()
	if err != nil {
		c := strings.Join(cmd.Args, " ")
		return false, fmt.Errorf("isMountedFFS: błąd polecenia: %q: %s", c, err)
	}

	buf := bytes.NewBuffer(out)
	scanner := bufio.NewScanner(buf)
	prefix := dev + " on " + dir
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("isMountedFFS: %s", err)
	}

	return false, nil
}

// fsck sprawdza filesystem na dysku disk. Argument disk ma postać
// duid.part.
func fsck(disk string) error {
	dev, err := diskNameFull(disk)
	if err == errNoDisk {
		return fmt.Errorf("fsck: dysk nie podłączony: %s", disk)
	}
	if err != nil {
		return fmt.Errorf("fsck: %s", err)
	}

	// polecenie: fsck -p /dev/sd2a
	cmd := exec.Command("fsck", "-p", dev)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		c := strings.Join(cmd.Args, " ")
		return fmt.Errorf("fsck: błąd polecenia: %q: %s", c, err)
	}

	return nil
}

func mount(disk0, disk1 string, dir string) error {
	mounted, err := isMountedSoftraid(disk0)
	if err != nil {
		return err
	}

	if mounted {
		log.Printf("dysk %q już jest podłączony do softraid", disk0)
	} else {
		err := mountSoftraid(disk0)
		if err != nil {
			return err
		}
	}

	mounted, err = isMountedFFS(disk1, dir)
	if err != nil {
		return err
	}

	if mounted {
		log.Printf("filesystem %q już jest podmontowany do %q", disk1, dir)
	} else {
		err = fsck(disk1)
		if err != nil {
			return err
		}
		err = mountFFS(disk1, dir)
		if err != nil {
			return err
		}
	}

	return nil
}

func unmount(disk0, disk1 string, dir string) error {
	mounted, err := isMountedFFS(disk1, dir)
	if err != nil {
		return err
	}

	if mounted {
		err := unmountFFS(disk1)
		if err != nil {
			return err
		}
	} else {
		log.Printf("filesystem %q już jest odmontowany od %q", disk1, dir)
	}

	mounted, err = isMountedSoftraid(disk0)
	if err != nil {
		return err
	}

	if mounted {
		err := unmountSoftraid(disk1)
		if err != nil {
			return err
		}
	} else {
		log.Printf("dysk %q już jest odłączony od softraid", disk0)
	}

	return nil
}

func hexdigit(c rune) bool {
	if '0' <= c && c <= '9' {
		return true
	}
	if 'a' <= c && c <= 'f' {
		return true
	}
	return false
}

func validate(disk string) error {
	a := strings.Split(disk, ".")
	if len(a) < 2 {
		return errors.New("brak specyfikacji partycji: brak znaku '.'")
	}

	if len(a) > 2 {
		return errors.New("zła specyfikacja partycji: za dużo znaków '.'")
	}

	if len(a[0]) != 16 {
		return errors.New("zła długość DUID")
	}

	for _, c := range a[0] {
		if !hexdigit(c) {
			return errors.New("DUID zawiera znaki nie 'hexdigit'")
		}
	}

	if len(a[1]) != 1 {
		return errors.New("zła długość specyfikacji partycji")
	}

	return nil
}

func main() {
	log.SetPrefix("cryptmount: ")
	log.SetFlags(0)

	flag.BoolVar(hFlag, "help", false, "wyświetl help")
	flag.Usage = usage
	flag.Parse()

	if *hFlag {
		help()
	}

	if err := validate(*disk0); err != nil {
		fmt.Fprintf(os.Stderr, "zła wartość opcji -disk0: %q: %s\n", *disk0, err)
		usage()
	}
	if err := validate(*disk1); err != nil {
		fmt.Fprintf(os.Stderr, "zła wartość opcji -disk1: %q: %s\n", *disk1, err)
		usage()
	}

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "brak argumentu dir")
		usage()
	}
	dir := flag.Arg(0)

	if *uFlag {
		err := unmount(*disk0, *disk1, dir)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err := mount(*disk0, *disk1, dir)
	if err != nil {
		log.Fatal(err)
	}
}
