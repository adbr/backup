// 2015-02-10 Adam Bryt

/*
Program cryptmount podłącza szyfrowaną partycję do softraid a
następnie montuje partycję na szyfrowanym dysku do filesystemu.

Jest przeznaczony dla systemu OpenBSD. Do obsługi szyfrowanej
partycji jest używany softraid(4) i polecenie bioctl(8).

Celem programu jest ułatwienie montowania dysku przy użyciu jego
DUID, który jest stały w przeciwieństwie do nazw z /dev.  Ponieważ
polecenie bioctl musi mieć nazwę dysku z /dev, program tłumaczy
nazwy DUID na aktualne nazwy dysków w katalogu /dev używając
polecenia 'sysctl hw.disknames'.

Sposób użycia:

     cryptmount [flags] -disk0=diskspec -disk1=diskspec -dir=dir

Wartości opcji disk0 i disk1 (diskspec) mają format 'DUID.PART',
gdzie DUID jest unikalnym identyfikatorem dysku z disklabel, a PART
jest pojedynczą literą oznaczającą partycję na tym dysku (np.
"a3a6acb427840bc0.a").

Opcja -disk0 specyfikuje zaszyfrowaną partycję typu RAID na dysku
fizycznym (np. USB), która zostanie podłączona do softraid0.
Opcja -disk1 specyfikuje partycję typu FFS na rozszyfrowanym dysku
logicznym, która zostanie podmontowana w katologu dir.

Opcje:

     -disk0=""
        dysk i partycja zaszyfrowana typu RAID do podłączenie do
        softraid (opcja wymagana)
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
     	wyświetl help
*/
package main
