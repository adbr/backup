// 2015-02-10 Adam Bryt

/*
Program snapshot kopiuje katalog przy użyciu polecenia rsync(1). Tworzy
kolejne snapshoty w katalogach o nazwach typu '2015-02-10T18:07:39'.
Jeśli pliki w aktualnym snapshocie nie zmieniły się, to nie są
kopiowane tylko są tworzone hard linki do plików w poprzednim
snapshocie.

Sposób użycia:
	snapshot [opcje] -src=filesystem -dst=directory
Opcje:
	-src filesystem
		backupowany filesystem
	-dst directory
		docelowy katalog z backupami
	-exclude string
		lista wzorców ignorowanych plików "pattern,pattern,..."
		(domyślnie: "")
	-logfile filename
		plik z logami (domyślnie: "")
	-rsync filename
		nazwa polecenia rsync (domyślnie: "rsync")
	-rsyncopts string
		opcje polecenia rsync (domyślnie: "-avxH8")
	-h	sposób użycia
	-help	dokumentacja

Do kopiowania jest używane polecenie rsync(1) z następującymi opcjami:

	-a			archive mode
	-v			verbose
	-x			don't cross filesystem boundaries
	-H			preserve hard links
	-8			8-bit output
	--link-dest=DIR		hardlink to files in DIR when unchanged
	--exclude=PATTERN	exclude files matching PATTERN
*/
package main
