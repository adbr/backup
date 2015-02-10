# backup

Narzędzia do robienia backupu typu snapshot z użyciem rsync.

- cryptmount - Montuje szyfrowany dysk. Narzędzie dla systemu OpenBSD. 
- snapshot - Kopiuje katalog przy użyciu rsync. Tworzy kolejne snapshoty w
  katalogach o nazwach typu '2015-02-10T18:07:39'.
- examples - przykładowe skrypty

## Instalacja:

- Zainstalować rsync

- Skompilować program cryptmount i skopiować go do katalogu bin. (opcjonalnie -
  tylko dla OpenBSD):

		$ cd cryptmount
		$ go build
		$ go install

- Skompilować program snapshot i skopiować go do katalogu bib:

		$ cd snapshot
		$ go build
		$ go install

- Skopiować skrypty z examples do katalogu bin i zmodyfikować je.
