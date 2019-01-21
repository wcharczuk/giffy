// Package db provides a basic abstraction layer above normal database/sql that makes it easier to
// interact with the database and organize database related code. It is not intended to replace actual sql
// (you write queries yourself in sql). It also includes some helpers to organize creating a connection
// to a database from a config file or object.
package db
