/**
 * 功能：storage.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package storage

import "io"

// Storage defines the interface for file storage operations
type Storage interface {
	Upload(filename string, content io.Reader) (string, error)
	GetURL(filename string) (string, error)
	Delete(filename string) error
}

var GlobalStorage Storage
