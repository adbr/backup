// 2015-02-12 Adam Bryt

/*
Program logrotate służy do rotacji plików z logami. Jeśli rozmiar
pliku jest większy niż wartość podana w opcji -size to plik jest
archiwizowany: nazwa pliku log jest zmieniana na log.0 i tworzony jest
pusty plik log.

Jeśli istnieją wcześniej zarchiwizowane pliki z logami to ich numery
są zwiększane, np:

	log.0 -> log.1
	log.1 -> log.2
	...

Maksymalną liczbę zarchiwizowanych plików z logami określa opcja -num.

Sposób użycia:
	logrotate [opcje] logfile
Opcje:
	-num int
		maksymalna liczba archiwizowanych plików (domyślnie:
		0, czyli bez ograniczenia)
	-size int
		wielkość (w bajtach) archiwizowanego pliku, jeśli
		rozmiar pliku logfile jest większy niż -size to plik
		jest archiwizowany (domyślnie: 0, czyli bez
		ograniczenia)
	-v	wyświetlanie komunikatów (verbose)
	-h	sposób użycia
	-help	dokumentacja
*/
package main
