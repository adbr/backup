// 2015-02-10 Adam Bryt

/*
Program snapshot kopiuje katalog przy użyciu polecenia rsync(1). Tworzy
kolejne snapshoty w katalogach o nazwach typu '2015-02-10T18:07:39'.
Jeśli pliki w aktualnym snapshocie nie zmieniły się, to nie są
kopiowane tylko są tworzone hard linki do plików w poprzednim
snapshocie.

Sposób użycia:
	snapshot [opcje] -fs=filesystem -dest=directory
Opcje:
	-fs filesystem
		backupowany filesystem
	-dest directory
		katalog docelowy
	-exclude string
		lista wzorców ignorowanych plików "pattern,pattern,..."
		(domyślnie: "")
	-logfile filename
		plik z logami (domyślnie: "")
	-rsync filename
		nazwa polecenia rsync (domyślnie: "rsync")
	-rsync_flags string
		opcje polecenia rsync (domyślnie: "-avxH8")

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
