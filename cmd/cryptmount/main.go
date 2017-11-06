// 2017-10-24 adbr

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/adbr/backup/cryptmount"
)

func main() {
	log.SetPrefix("cryptmount: ")
	log.SetFlags(0)

	// Flagi programu - argumenty usage są puste ("") ponieważ
	// używam własnej funkcji Usage.
	disk0 := flag.String("disk0", "", "")
	disk1 := flag.String("disk1", "", "")
	dir := flag.String("dir", "", "")
	mountopts := flag.String("mountopts", "-o softdep", "")
	u := flag.Bool("u", false, "")
	h := flag.Bool("h", false, "")
	help := flag.Bool("help", false, "")

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
	if *disk0 == "" {
		log.Print("brak opcji -disk0")
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}
	if *disk1 == "" {
		log.Print("brak opcji -disk1")
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}
	if *dir == "" {
		log.Print("brak opcji -dir")
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}

	if *u {
		err := unmount(*disk1)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	err := mount(*disk0, *disk1, *dir, *mountopts)
	if err != nil {
		log.Fatal(err)
	}
}

// mount montuje disk0 do softraid i disk1 do filesystemu w punkcie
// dir i z opcjami mountopts.
func mount(disk0, disk1, dir, mountopts string) error {
	err := cryptmount.MountSoftraid(disk0)
	if err != nil {
		return err
	}
	err = cryptmount.Fsck(disk1)
	if err != nil {
		return err
	}
	err = cryptmount.MountFS(disk1, dir, mountopts)
	if err != nil {
		return err
	}
	return nil
}

// unmount odmontowuje disk z filesystemu i z softraidu. Argument disk
// jest dyskiem logicznym softraidu.
func unmount(disk string) error {
	err := cryptmount.UnmountFS(disk)
	if err != nil {
		return err
	}
	err = cryptmount.UnmountSoftraid(disk)
	if err != nil {
		return err
	}
	return nil
}

// Stała usageText zawiera opis opcji programu wyświetlany przy użyciu
// opcji -h lub w przypadku błędu parsowania opcji.
const usageText = `Sposób użycia:
	cryptmount [opcje] -disk0=diskspec -disk1=diskspec -dir=directory
Opcje:
	-disk0 diskspec
		dysk i partycja (DUID.PART) typu RAID, zaszyfrowana,
		do podłączenie do softraid
	-disk1 diskspec
		dysk i partycja (DUID.PART) typu FFS na rozszyfrowanym
		dysku logicznym, do podmontowania w katalogu directory
	-dir directory
		katalog do podmontowania filesystemu
	-mountopts string
		opcje dla polecenia mount (domyślnie: "-o softdep")
	-u	odmontuj dyski (unmount)
	-h	sposób użycia
	-help	dokumentacja
`

// Stała helpText zawiera opis programu wyświetlany przy użyciu opcji
// -help. Treść jest identyczna jak w doc comment programu z pliku
// doc.go.
const helpText = `
Program cryptmount podłącza szyfrowaną partycję do softraid, a
następnie montuje partycję na szyfrowanym dysku do filesystemu.

Jest przeznaczony dla systemu OpenBSD. Do obsługi szyfrowanej partycji
jest używany softraid(4) i polecenie bioctl(8). Celem programu jest
ułatwienie montowania dysku przy użyciu jego DUID, który jest stały w
przeciwieństwie do nazw z /dev.

Sposób użycia:

	cryptmount [opcje] -disk0=diskspec -disk1=diskspec -dir=directory

Wartości opcji disk0 i disk1 (diskspec) mają format 'DUID.PART', gdzie
DUID jest unikalnym identyfikatorem dysku z disklabel, a PART jest
pojedynczą literą oznaczającą partycję na tym dysku (np.
"a3a6acb427840bc0.a").

Opcja -disk0 specyfikuje zaszyfrowaną partycję typu RAID na dysku
fizycznym (np. USB), która zostanie podłączona do softraid0.  Opcja
-disk1 specyfikuje partycję typu FFS na rozszyfrowanym dysku
logicznym, która zostanie podmontowana w katalogu directory.

Opcje:

	-disk0 diskspec
		dysk i partycja (DUID.PART) typu RAID, zaszyfrowana,
		do podłączenie do softraid
	-disk1 diskspec
		dysk i partycja (DUID.PART) typu FFS na rozszyfrowanym
		dysku logicznym, do podmontowania w katalogu directory
	-dir directory
		katalog do podmontowania filesystemu
	-mountopts string
		opcje dla polecenia mount (domyślnie: "-o softdep")
	-u	odmontuj dyski (unmount)
	-h	sposób użycia
	-help	dokumentacja
`
