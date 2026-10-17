package main

// Base de test : module sqlite3 de Python (voir testdb_linux.go).
func openTestDB(path string) (DB, error) { return openPyDB(path) }
