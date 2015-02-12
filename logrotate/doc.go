// 2015-02-12 Adam Bryt

/*
Program logrotate służy do rotacji plików z logami. Jeśli rozmiar pliku
jest większy niż wartość podana w opcji -size to plik jest archiwizowany:
nazwa pliku log jest zmieniana na log.0 i tworzony jest pusty plik log.
*/
package main
