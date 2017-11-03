// 2015-02-10 Adam Bryt

/*
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
*/
package main
