// 2017-10-24 adbr

// Pakiet cryptmount dostarcza funkcje do montowania dysku
// szyfrowanego do softraid i filesystemu na szyfrowanym
// dysku. Funkcje używają poleceń systemowych (bioctl, mount, umount,
// fsck).
package cryptmount

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// MountSoftraid podłącza zaszyfrowany dysk do softraid. Argument disk
// ma postać DUID.PART, gdzie DUID jest unikalnym identyfikatorem
// dysku, a PART jest pojedynczą literą oznaczającą partycją typu RAID
// na tym dysku (np.: "a3a6acb427840bc0.a"). Funkcja używa polecenia
// systemowego bioctl oraz stdin/stdout do wczytania passphrase w celu
// rozszyfrowania dysku.
func MountSoftraid(disk string) error {
	cmd := exec.Command("/sbin/bioctl", "-c", "C", "-l", disk, "softraid0")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("executing: %q", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("bioctl: %s", err)
	}
	return nil
}

// UnmountSoftraid odłącza szyfrowany dysk od softraid. Argument disk
// ma postać DUID.PART, gdzie DUID jest unikalnym identyfikatorem
// dysku, a PART jest pojedynczą literą oznaczającą partycją (np.:
// "a3a6acb427840bc0.a"). Argument disk to szyfrowany dysk logiczny,
// czyli dysk, który pojawił się po podłączeniu partycji RAID do
// softraid. Partycja w specyfikacji dysku nie ma znaczenia. Funkcja
// używa polecenia systemowego bioctl.
func UnmountSoftraid(disk string) error {
	cmd := exec.Command("/sbin/bioctl", "-d", disk)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("executing: %q", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("bioctl: %s", err)
	}
	return nil
}

// MountFS montuje filesystem disk w katalogu dir z opcjami
// options. Argument disk ma postać DUID.PART, gdzie DUID jest
// unikalnym identyfikatorem dysku, a PART jest pojedynczą literą
// oznaczającą partycją (np.: "a3a6acb427840bc0.a"). Funkcja używa
// polecenia systemowego mount. Argument options ma postać opcji
// polecenia mount, np.: "-o softdep".
func MountFS(disk, dir, options string) error {
	var args []string
	args = append(args, strings.Fields(options)...)
	args = append(args, disk, dir)
	cmd := exec.Command("/sbin/mount", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("executing: %q", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("mount: %s", err)
	}
	return nil
}

// UnmountFS odmontowuje filesystem disk.  Argument disk ma postać
// DUID.PART, gdzie DUID jest unikalnym identyfikatorem dysku, a PART
// jest pojedynczą literą oznaczającą partycją (np.:
// "a3a6acb427840bc0.a"). Używa polecenia systemowego umount.
func UnmountFS(disk string) error {
	cmd := exec.Command("/sbin/umount", disk)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("executing: %q", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("umount: %s", err)
	}
	return nil
}

// Fsck sprawdza filesystem disk.  Argument disk ma postać DUID.PART,
// gdzie DUID jest unikalnym identyfikatorem dysku, a PART jest
// pojedynczą literą oznaczającą partycją (np.:
// "a3a6acb427840bc0.a"). Używa polecenia systemowego 'fsck -p'.
func Fsck(disk string) error {
	cmd := exec.Command("/sbin/fsck", "-p", disk)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("executing: %q", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("fsck: %s", err)
	}
	return nil
}
