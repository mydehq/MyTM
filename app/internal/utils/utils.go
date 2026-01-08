// Package utils provides common helper functions used across the application.
package utils

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileExists checks if a file exists and is not a directory.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	// It exists, but is it a file?
	return !info.IsDir()
}

// DirExists checks if a path exists and is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// GenerateHash calculates the checksum of a file.
// It supports multiple algorithms like md5, sha1, sha256.
func GenerateHash(filePath, algo string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close() // Ensure file is closed even if we return early

	var hasher hash.Hash

	// Factory pattern: create the right hasher based on string input
	switch strings.ToLower(algo) {
	case "md5":
		hasher = md5.New()
	case "sha1":
		hasher = sha1.New()
	case "sha256":
		hasher = sha256.New()
	case "sha512":
		hasher = sha512.New()
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algo)
	}

	// io.Copy efficiently streams the file content into the hasher.
	// It doesn't load the whole file into RAM.
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	// Get final hash bytes and encode to hex string
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// CopyFile copies a file from source to destination.
// It uses io.Copy for efficiency.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// CreateTarGz recursively archives a directory into a .tar.gz file.
// This function demonstrates complex file I/O operations (Walk, Symlinks, Tar Header, Gzip).
func CreateTarGz(srcDir, destFile string) error {
	// 1. Create the output file
	file, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// 2. Wrap output file in Gzip Writer
	gw := gzip.NewWriter(file)
	defer gw.Close()

	// 3. Wrap Gzip Writer in Tar Writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// 4. Walk the directory tree recursively
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Handle Symlinks: We need to know where they point to.
		var linkTarget string
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}

		// Create a Tar Header from the file info
		header, err := tar.FileInfoHeader(info, linkTarget)
		if err != nil {
			return err
		}

		// Use relative path for header name to preserve structure inside the archive
		// e.g. /home/user/mytm/MyTM/themes/dracula/src/main.css -> src/main.css
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Force PAX format: Basic USTAR format only supports 100 character filenames.
		// PAX supports arbitrary length paths (important for nested node_modules etc).
		header.Format = tar.FormatPAX

		// Ensure forward slashes for tar compatibility across OS (Windows uses backslash)
		relPath = filepath.ToSlash(relPath)

		// Don't include the root directory itself as an entry (./)
		if relPath == "." {
			return nil
		}

		header.Name = relPath

		// Write Header
		if err := tw.WriteHeader(header); err != nil {
			// fmt.Printf("FAILED writing header for '%s' (len=%d). Error: %v\n", relPath, len(relPath), err)
			return err
		}

		// If it's a regular file, write its content
		if !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}
		return nil
	})
}
