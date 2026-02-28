// Package fs provides a filesystem abstraction used throughout the application.
// Wrapping filesystem calls behind an interface makes unit-testing easier
// without requiring real disk access.
package fs

import (
	"io/fs"
	"os"
)

// FS is a minimal filesystem interface covering operations used by this tool.
type FS interface {
	Stat(name string) (fs.FileInfo, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
	Remove(name string) error
	RemoveAll(name string) error
}

// OsFS is the production implementation backed by the real os package.
type OsFS struct{}

func (OsFS) Stat(name string) (fs.FileInfo, error)      { return os.Stat(name) }
func (OsFS) ReadDir(name string) ([]fs.DirEntry, error) { return os.ReadDir(name) }
func (OsFS) ReadFile(name string) ([]byte, error)       { return os.ReadFile(name) }
func (OsFS) Remove(name string) error                   { return os.Remove(name) }
func (OsFS) RemoveAll(name string) error                { return os.RemoveAll(name) }
