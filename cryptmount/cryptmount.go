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
//	cryptmount [flags] -disk0=diskspec -disk1=diskspec -dir=dir
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
//	-dir=""
//		katalog do podmontowania filesystemu (opcja wymagana)
//	-mountopts="-o softdep"
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
	dir     = flag.String("dir", "", "katalog do podmontowania fs")
	mntopts = flag.String("mountopts", "-o softdep", "opcje polecenia mount")
	uFlag   = flag.Bool("u", false, "odmontuj dyski (unmount)")
	hFlag   bool // help flag
)

func init() {
	// zmiana domyślnej obsługi flag -h i -help
	flag.BoolVar(&hFlag, "h", false, "wyświetl help")
	flag.BoolVar(&hFlag, "help", false, "wyświetl help")
}

const usageStr = `usage: cryptmount [flags] -disk0=diskspec -disk1=diskspec -dir=dir
	-disk0=""
		dysk i partycja zaszyfrowana typu RAID do podłączenie
		do softraid (opcja wymagana)
	-disk1=""
		dysk i partycja typu FFS na rozszyfrowanym dysku logicznym
		do podmontowania w katalogu dir (opcja wymagana)
	-dir=""
		katalog do podmontowania filesystemu (opcja wymagana)
	-mountopts="-o softdep"
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
	os.Exit(1)
}

func help() {
	fmt.Println(usageStr)
	fmt.Println("")
	fmt.Println(helpStr)
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

// diskNameFull zwraca nazwę pliku specjalnego w /dev odpowiadającą dyskowi
// disk. Parametr disk ma postać duid.part. Np. dla disk 'e072adf1dcc1be16.a'
// zwróci '/dev/sd0a'. Jeśli dysk nie istnieje to zwraca errNoDisk.
func diskNameFull(disk string) (string, error) {
	a := strings.Split(disk, ".")
	duid, part := a[0], a[1]

	dev, err := diskName(duid)
	if err != nil {
		return "", err
	}

	name := "/dev/" + dev + part
	return name, nil
}

// isMountedSoftraid sprawdza czy dysk jest podpięty do softraid. Jeśli tak
// to zwraca true i dodatkowo nazwę dysku logicznego (rozszyfrowanego).
func isMountedSoftraid() (bool, string, error) {
	a := strings.Split(*disk0, ".")
	duid, part := a[0], a[1]

	dev, err := diskName(duid)
	if err == errNoDisk {
		return false, "", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("isMountedSoftraid: %s", err)
	}

	// uruchom polecenie 'bioctl softraid0'
	cmd := exec.Command("bioctl", "softraid0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s", out)
		c := strings.Join(cmd.Args, " ")
		return false, "", fmt.Errorf("isMountedSoftraid: błąd polecenia %q: %s", c, err)
	}

	dev = "<" + dev + part + ">" // nazwa dysku fizycznego
	var ldev string              // nazwa dysku logicznego (rozszyfrowanego)
	buf := bytes.NewBuffer(out)
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		line := scanner.Text()
		a := strings.Fields(line)
		if len(a) == 6 && a[5] == "CRYPTO" {
			ldev = a[4]
		}
		if len(a) == 6 && a[5] == dev {
			return true, ldev, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, "", fmt.Errorf("isMountedSoftraid: %s", err)
	}

	return false, "", nil
}

// isMountedFFS sprawdza czy filesystem na dysku logicznym (rozszyfrowanym)
// jest podmontowany w katalogu dir.
func isMountedFFS() (bool, error) {
	dev, err := diskNameFull(*disk1)
	if err == errNoDisk {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("isMountedFFS: %s", err)
	}

	// uruchomienie polecenia mount, wynik polecenia ma postać:
	//    /dev/sd1l on /home type ffs (local, nodev, nosuid, softdep)
	cmd := exec.Command("mount")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s", out)
		c := strings.Join(cmd.Args, " ")
		return false, fmt.Errorf("isMountedFFS: błąd polecenia: %q: %s", c, err)
	}

	// sprawdzenie czy wynik polecenia mount zawiera szukany filesystem
	buf := bytes.NewBuffer(out)
	scanner := bufio.NewScanner(buf)
	prefix := dev + " on " + *dir // np. '/dev/sd3d on /backup'
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

// Funckcja mountSoftraid podłącza szyfrowany dysk do softraid.
func mountSoftraid() error {
	ok, _, err := isMountedSoftraid()
	if err != nil {
		return fmt.Errorf("mount softraid: %s", err)
	}
	if ok {
		log.Printf("WARNING: mount softraid: dysk już jest podłączony do softraid: %q", *disk0)
		return nil
	}

	dev, err := diskNameFull(*disk0)
	if err == errNoDisk {
		return fmt.Errorf("mount softraid: dysk nie podłączony: %q", *disk0)
	}
	if err != nil {
		return fmt.Errorf("mount softraid: %s", err)
	}

	// uruchom polecenie bioctl
	// np. 'bioctl -c C -l /dev/sd2a softraid0'
	cmd := exec.Command("bioctl", "-c", "C", "-l", dev, "softraid0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmdstr := strings.Join(cmd.Args, " ")
	log.Printf("mount softraid: '%s'", cmdstr)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("mount softraid: błąd polecenia: %q: %s", cmdstr, err)
	}

	return nil
}

// Funkcja unmountSoftraid odłącza szyfrowany dysk od softraid.
func unmountSoftraid() error {
	ok, dev, err := isMountedSoftraid()
	if err != nil {
		return fmt.Errorf("unmount softraid: %s", err)
	}
	if !ok {
		log.Printf("WARNING: unmount softraid: dysk już jest odłączony od softraid: %q", *disk0)
		return nil
	}

	// uruchom polecenie bioctl
	// np. 'bioctl -d sd3'
	cmd := exec.Command("bioctl", "-d", dev)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmdstr := strings.Join(cmd.Args, " ")
	log.Printf("unmount softraid: '%s'", cmdstr)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("unmount softraid: błąd polecenia: %q: %s", cmdstr, err)
	}

	return nil
}

// fsck sprawdza filesystem na dysku logicznym (rozszyfrowanym).
func fsck() error {
	ok, err := isMountedFFS()
	if err != nil {
		return fmt.Errorf("fsck: %s", err)
	}
	if ok {
		log.Printf("WARNING: fsck: dysk %q jest podmontowany - nie sprawdzam", *disk1)
		return nil
	}

	dev, err := diskNameFull(*disk1)
	if err == errNoDisk {
		return fmt.Errorf("fsck: dysk nie podłączony: %s", *disk1)
	}
	if err != nil {
		return fmt.Errorf("fsck: %s", err)
	}

	// uruchomienie polecenia fsck
	// np. 'fsck -p /dev/sd2a'
	cmd := exec.Command("fsck", "-p", dev)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmdstr := strings.Join(cmd.Args, " ")
	log.Printf("fsck: '%s'", cmdstr)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("fsck: błąd polecenia: %q: %s", cmdstr, err)
	}

	return nil
}

// mountFFS montuje filesystem na dysku logicznym (rozszyfrowanym).
func mountFFS() error {
	ok, err := isMountedFFS()
	if err != nil {
		return fmt.Errorf("mount filesystem: %s", err)
	}
	if ok {
		log.Printf("WARNING: mount filesystem: dysk już jest podmontowany: %q", *disk1)
		return nil
	}

	dev, err := diskNameFull(*disk1)
	if err == errNoDisk {
		return fmt.Errorf("mount filesystem: dysk nie podłączony: %s", *disk1)
	}
	if err != nil {
		return fmt.Errorf("mount filesystem: %s", err)
	}

	// uruchomienie polecenia mount
	// np. 'mount -o softdep /dev/sd2a /backup'
	opts := strings.Fields(*mntopts)
	args := []string{}
	args = append(args, opts...)
	args = append(args, dev, *dir)
	cmd := exec.Command("mount", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmdstr := strings.Join(cmd.Args, " ")
	log.Printf("mount filesystem: '%s'", cmdstr)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("mount filesystem: błąd polecenia: %q: %s", cmdstr, err)
	}

	return nil
}

// unmountFFS odmontowuje filesystem na dysku logicznym (rozszyfrowanym).
func unmountFFS() error {
	ok, err := isMountedFFS()
	if err != nil {
		return fmt.Errorf("unmount filesystem: %s", err)
	}
	if !ok {
		log.Printf("WARNING: unmount filesystem: dysk już jest odmontowany: %q", *disk1)
		return nil
	}

	dev, err := diskNameFull(*disk1)
	if err == errNoDisk {
		return fmt.Errorf("unmount filesystem: dysk nie podłączony: %s", *disk1)
	}
	if err != nil {
		return fmt.Errorf("unmount filesystem: %s", err)
	}

	// uruchomienie polecenia umount
	// np. 'umount /dev/sd2a'
	cmd := exec.Command("umount", dev)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmdstr := strings.Join(cmd.Args, " ")
	log.Printf("unmount filesystem: '%s'", cmdstr)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("unmount filesystem: błąd polecenia: %q: %s", cmdstr, err)
	}

	return nil
}

func mount() error {
	err := mountSoftraid()
	if err != nil {
		return err
	}

	err = fsck()
	if err != nil {
		return err
	}

	err = mountFFS()
	if err != nil {
		return err
	}

	return nil
}

func unmount() error {
	err := unmountFFS()
	if err != nil {
		return err
	}

	err = unmountSoftraid()
	if err != nil {
		return err
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

// validate sprawdza poprawność specyfikacji dysku i partycji.
// Parametr disk ma postać duid.part.
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

	flag.Usage = usage
	flag.Parse()

	if hFlag {
		help()
	}

	if err := validate(*disk0); err != nil {
		fmt.Fprintf(os.Stderr,
			"zła wartość opcji -disk0: %q: %s\n", *disk0, err)
		usage()
	}
	if err := validate(*disk1); err != nil {
		fmt.Fprintf(os.Stderr,
			"zła wartość opcji -disk1: %q: %s\n", *disk1, err)
		usage()
	}
	if *dir == "" {
		fmt.Fprintf(os.Stderr, "zła wartość opcji -dir: %q\n", *dir)
		usage()
	}

	if *uFlag {
		err := unmount()
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		return
	}

	err := mount()
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
}
