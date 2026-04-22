// Package compression provides utilities for working with compressed archives.
// Supports ZIP and TAR archive creation and extraction with various compression methods.
//
// # ZIP Usage
//
// Extracting:
//
//	err := compression.UnZip("/path/to/archive.zip", "/path/to/destination")
//	if err != nil {
//	    log.Fatalf("Failed to extract: %v", err)
//	}
//
// Creating:
//
//	// With deflate compression (default)
//	err := compression.Zip("/path/to/source", "/path/to/archive.zip")
//
//	// Without compression (store only)
//	err := compression.Zip("/path/to/source", "/path/to/archive.zip",
//	    compression.WithZipStore())
//
// # TAR Usage
//
// Extracting (auto-detects compression):
//
//	err := compression.UnTar("/path/to/archive.tar.gz", "/path/to/destination")
//	if err != nil {
//	    log.Fatalf("Failed to extract: %v", err)
//	}
//
// Creating:
//
//	// Gzip compression (auto-detected from extension)
//	err := compression.Tar("/path/to/source", "/path/to/archive.tar.gz")
//
//	// Explicit compression option
//	err := compression.Tar("/path/to/source", "/path/to/archive.tar",
//	    compression.WithBzip2Compression())
//
// # Supported Formats
//
//   - ZIP: .zip (store or deflate compression)
//   - TAR: .tar, .tar.gz/.tgz, .tar.bz2/.tbz2
//
// # Security
//
// All extraction functions include protection against path traversal attacks
// (zip slip/tar slip) by validating that extracted files remain within the
// destination directory.
package compression

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ErrZipSlip is returned when a zip archive contains files that would be
// extracted outside the destination directory (path traversal attack).
var ErrZipSlip = errors.New("zip slip: file path escapes destination directory")

// UnZip extracts a ZIP archive to the specified destination directory.
// It creates directories as needed and preserves the archive's directory structure.
//
// Security: This function validates that all extracted files remain within the
// destination directory to prevent zip slip (path traversal) attacks.
//
// Parameters:
//   - zipPath: Path to the ZIP file to extract
//   - destDir: Destination directory for extracted files
//
// Returns an error if the archive cannot be opened, read, if file operations fail,
// or if a zip slip attack is detected.
func UnZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// Clean and resolve the destination directory for consistent comparison
	destDir = filepath.Clean(destDir)

	for _, f := range r.File {
		if err := extractZipFile(f, destDir); err != nil {
			return err
		}
	}
	return nil
}

// extractZipFile extracts a single file from a zip archive with zip slip protection.
func extractZipFile(f *zip.File, destDir string) error {
	// Clean the file path and join with destination
	fPath := filepath.Join(destDir, filepath.Clean(f.Name))

	// Security check: ensure the resolved path is within the destination directory
	// This prevents zip slip attacks where malicious archives contain paths like
	// "../../../etc/passwd" that would write outside the intended directory
	if !strings.HasPrefix(fPath, destDir+string(os.PathSeparator)) && fPath != destDir {
		return ErrZipSlip
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(fPath, os.ModePerm)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
		return err
	}

	return extractFileContent(f, fPath)
}

// extractFileContent extracts the content of a zip file entry to the given path.
func extractFileContent(f *zip.File, fPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	outFile, err := os.Create(fPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	return err
}

// ZipOption is a function that configures zip creation options.
type ZipOption func(*zipConfig)

// zipConfig holds configuration for zip creation.
type zipConfig struct {
	compressionMethod uint16
}

// WithZipDeflate configures zip creation to use deflate (gzip) compression.
// This is the standard compression method for ZIP archives.
func WithZipDeflate() ZipOption {
	return func(c *zipConfig) {
		c.compressionMethod = zip.Deflate
	}
}

// WithZipStore configures zip creation to use no compression (store only).
// Files are stored as-is without any compression.
func WithZipStore() ZipOption {
	return func(c *zipConfig) {
		c.compressionMethod = zip.Store
	}
}

// Zip creates a ZIP archive from the specified source path.
// The source can be a single file or a directory (which will be recursively archived).
//
// By default, deflate (gzip) compression is used. This can be changed using ZipOption functions:
//   - WithZipDeflate() - Use deflate compression (default)
//   - WithZipStore() - Store files without compression
//
// The archive preserves file permissions, modification times, and directory structure.
//
// Note: Bzip2 compression is not supported for ZIP archives due to limited
// Go standard library support.
//
// Parameters:
//   - sourcePath: Path to the file or directory to archive
//   - destZipPath: Destination path for the zip archive
//   - opts: Optional configuration functions
//
// Returns an error if the source cannot be read, the destination cannot be created,
// or if file operations fail during archiving.
//
// Example:
//
//	// Create zip with deflate compression (default)
//	err := Zip("/path/to/source", "/path/to/archive.zip")
//
//	// Create zip without compression
//	err := Zip("/path/to/source", "/path/to/archive.zip", WithZipStore())
func Zip(sourcePath, destZipPath string, opts ...ZipOption) error {
	// Configure compression (default to deflate)
	config := &zipConfig{
		compressionMethod: zip.Deflate,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Create destination file
	file, err := os.Create(destZipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create zip writer
	zw := zip.NewWriter(file)
	defer zw.Close()

	// Get the base directory for calculating relative paths
	baseDir := filepath.Dir(sourcePath)

	// Add files to archive
	return addToZip(zw, sourcePath, baseDir, config.compressionMethod)
}

// addToZip recursively adds files and directories to a zip archive.
func addToZip(zw *zip.Writer, sourcePath, baseDir string, compressionMethod uint16) error {
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories (they're implicitly created by file paths)
		if info.IsDir() {
			return nil
		}

		// Create zip file header
		header, err := createZipFileHeader(info, path, baseDir, compressionMethod)
		if err != nil {
			return err
		}

		// Create writer for this file
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		// Write file content if it's a regular file
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(writer, file); err != nil {
				return err
			}
		}

		return nil
	})
}

// createZipFileHeader creates a zip file header from file info.
func createZipFileHeader(info os.FileInfo, path, baseDir string, compressionMethod uint16) (*zip.FileHeader, error) {
	// Create header from file info
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return nil, err
	}

	// Calculate relative path for the archive
	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		return nil, err
	}

	// Use forward slashes in zip archives (portable across platforms)
	header.Name = filepath.ToSlash(relPath)

	// Set compression method
	header.Method = compressionMethod

	return header, nil
}
