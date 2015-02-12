// 2015-02-12 Adam Bryt

/*
Program logrotate służy do rotacji plików z logami. Jeśli rozmiar pliku
jest większy niż wartość podana w opcji -size to plik jest archiwizowany:
nazwa pliku log jest zmieniana na log.0 i tworzony jest pusty plik log.

Jeśli istnieją wcześniej zarchiwizowane pliki z logami to ich numery 
są zwiększane, np:

	log.0 -> log.1
	log.1 -> log.2
	...

Maksymalną liczbę zarchiwizowanych plików z logami określa flaga -num.

Sposób użycia:

	logrotate [flags] logfile

Flagi:

	-num=0
		liczba archiwizowanych plików (0 bez ograniczeń)
	-size=0
		wielkość (w bajtach) archiwizowanego pliku (0 bez rotacji),
		jeśli rozmiar pliku logfile jest większy niż -size to plik
		jest archiwizowany
*/
package main
