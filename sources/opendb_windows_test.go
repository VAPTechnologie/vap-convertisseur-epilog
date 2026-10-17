package main

// Base de test : un fichier vide est une base SQLite valide, ouverte par winsqlite3.
func openTestDB(path string) (DB, error) { return openRW(path) }
