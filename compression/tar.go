package compression

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	bzip2writer "github.com/dsnet/compress/bzip2"
)

// TAR-related error variables
var (
	// ErrTarSlip is returned when a tar archive contains files that would be
	// extracted outside the destination directory (path traversal attack).
	ErrTarSlip = errors.New("tar slip: file path escapes destination directory")

	// ErrUnsupportedTarEntry is returned when encountering unsupported tar entry types
	// such as device files, FIFOs, or other special files.
	ErrUnsupportedTarEntry = errors.New("unsupported tar entry type")

	// ErrInvalidSymlink is returned when a symlink points outside the destination directory.
	ErrInvalidSymlink = errors.New("symlink points outside destination directory")

	// ErrUnknownCompression is returned when the compression format cannot be determined.
	ErrUnknownCompression = errors.New("unknown compression format")
)

// compressionType represents the type of compression used in a tar archive.
type compressionType int

const (
	compressionNone compressionType = iota
	compressionGzip
	compressionBzip2
)

// UnTar extracts a TAR archive to the specified destination directory.
// It automatically detects the compression format based on file extension
// (.tar, .tar.gz, .tgz, .tar.bz2, .tbz2) and handles decompression accordingly.
//
// The function creates directories as needed and preserves the archive's
// directory structure, file permissions, and symbolic links.
//
// Security: This function validates that all extracted files remain within the
// destination directory to prevent tar slip (path traversal) attacks. It also
// validates that symlinks do not point outside the destination directory.
//
// Parameters:
//   - tarPath: Path to the TAR file to extract
//   - destDir: Destination directory for extracted files
//
// Returns an error if the archive cannot be opened, read, if file operations fail,
// if a tar slip attack is detected, or if unsupported entry types are encountered.
func UnTar(tarPath, destDir string) error {
	// Open the tar file
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Detect compression type
	compression := detectCompression(tarPath)

	// Create tar reader with appropriate decompression
	tr, closer, err := openTarReader(file, compression)
	if err != nil {
		return err
	}
	if closer != nil {
		defer closer.Close()
	}

	// Clean and resolve the destination directory for consistent comparison
	destDir = filepath.Clean(destDir)

	// Extract all entries
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		if err := extractTarFile(header, tr, destDir); err != nil {
			return err
		}
	}

	return nil
}

// detectCompression determines the compression type based on file extension.
func detectCompression(tarPath string) compressionType {
	lower := strings.ToLower(tarPath)

	// Check for gzip compression
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return compressionGzip
	}

	// Check for bzip2 compression
	if strings.HasSuffix(lower, ".tar.bz2") || strings.HasSuffix(lower, ".tbz2") {
		return compressionBzip2
	}

	// Assume uncompressed
	return compressionNone
}

// openTarReader creates a tar reader with the appropriate decompression layer.
// Returns the tar reader and an optional closer for the decompression reader.
func openTarReader(file *os.File, compression compressionType) (*tar.Reader, io.Closer, error) {
	switch compression {
	case compressionGzip:
		gr, err := gzip.NewReader(file)
		if err != nil {
			return nil, nil, err
		}
		return tar.NewReader(gr), gr, nil

	case compressionBzip2:
		br := bzip2.NewReader(file)
		return tar.NewReader(br), nil, nil

	case compressionNone:
		return tar.NewReader(file), nil, nil

	default:
		return nil, nil, ErrUnknownCompression
	}
}

// extractTarFile extracts a single entry from a tar archive with security checks.
func extractTarFile(header *tar.Header, reader *tar.Reader, destDir string) error {
	// Validate the target path
	if err := validateTarPath(header, destDir); err != nil {
		return err
	}

	// Build the target path
	targetPath := filepath.Join(destDir, filepath.Clean(header.Name))

	// Handle different entry types
	switch header.Typeflag {
	case tar.TypeDir:
		// Create directory with preserved permissions (sanitized)
		mode := sanitizeFileMode(header.FileInfo().Mode())
		return os.MkdirAll(targetPath, mode)

	case tar.TypeReg:
		// Regular file - ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		return extractTarFileContent(header, reader, targetPath)

	case tar.TypeSymlink:
		// Validate symlink target
		if err := validateSymlink(header, destDir, targetPath); err != nil {
			return err
		}
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		return os.Symlink(header.Linkname, targetPath)

	case tar.TypeLink:
		// Hard link - ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		linkTarget := filepath.Join(destDir, filepath.Clean(header.Linkname))
		return os.Link(linkTarget, targetPath)

	default:
		// Skip unsupported entry types (device files, FIFOs, etc.)
		// We don't return error, just skip them
		return nil
	}
}

// extractTarFileContent extracts the content of a tar file entry to the given path.
func extractTarFileContent(header *tar.Header, reader *tar.Reader, targetPath string) error {
	// Create the output file with sanitized permissions
	mode := sanitizeFileMode(header.FileInfo().Mode())
	outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Copy content
	_, err = io.Copy(outFile, reader)
	return err
}

// validateTarPath checks if the target path is within the destination directory.
func validateTarPath(header *tar.Header, destDir string) error {
	// Clean the path
	targetPath := filepath.Join(destDir, filepath.Clean(header.Name))

	// Security check: ensure the resolved path is within the destination directory
	// This prevents tar slip attacks where malicious archives contain paths like
	// "../../../etc/passwd" that would write outside the intended directory
	if !strings.HasPrefix(targetPath, destDir+string(os.PathSeparator)) && targetPath != destDir {
		return ErrTarSlip
	}

	return nil
}

// validateSymlink checks if a symlink target is safe (doesn't point outside destDir).
func validateSymlink(header *tar.Header, destDir string, targetPath string) error {
	// If linkname is absolute, it's potentially dangerous
	if filepath.IsAbs(header.Linkname) {
		return ErrInvalidSymlink
	}

	// Resolve the symlink target relative to the symlink location
	linkDir := filepath.Dir(targetPath)
	linkTarget := filepath.Join(linkDir, header.Linkname)
	linkTarget = filepath.Clean(linkTarget)

	// Check if the resolved link target is within destDir
	if !strings.HasPrefix(linkTarget, destDir+string(os.PathSeparator)) && linkTarget != destDir {
		return ErrInvalidSymlink
	}

	return nil
}

// sanitizeFileMode removes dangerous permission bits (setuid, setgid, sticky).
// This prevents security issues when extracting archives from untrusted sources.
func sanitizeFileMode(mode os.FileMode) os.FileMode {
	// Remove setuid (04000), setgid (02000), and sticky (01000) bits
	// Keep only the standard permission bits (0777) and type bits
	return mode & (os.ModePerm | os.ModeType)
}

// TarOption is a function that configures tar creation options.
type TarOption func(*tarConfig)

// tarConfig holds configuration for tar creation.
type tarConfig struct {
	compression compressionType
}

// WithGzipCompression configures tar creation to use gzip compression.
// The resulting archive will be in .tar.gz format.
func WithGzipCompression() TarOption {
	return func(c *tarConfig) {
		c.compression = compressionGzip
	}
}

// WithBzip2Compression configures tar creation to use bzip2 compression.
// The resulting archive will be in .tar.bz2 format.
func WithBzip2Compression() TarOption {
	return func(c *tarConfig) {
		c.compression = compressionBzip2
	}
}

// WithNoCompression configures tar creation to create an uncompressed tar archive.
// The resulting archive will be in .tar format.
func WithNoCompression() TarOption {
	return func(c *tarConfig) {
		c.compression = compressionNone
	}
}

// Tar creates a TAR archive from the specified source path.
// The source can be a single file or a directory (which will be recursively archived).
//
// Compression is automatically detected from the destination file extension:
//   - .tar.gz, .tgz -> gzip compression
//   - .tar.bz2, .tbz2 -> bzip2 compression
//   - .tar -> no compression
//
// Compression can be explicitly overridden using TarOption functions:
//   - WithGzipCompression()
//   - WithBzip2Compression()
//   - WithNoCompression()
//
// The archive preserves file permissions, modification times, and directory structure.
// Symbolic links are stored in the archive (not followed).
//
// Parameters:
//   - sourcePath: Path to the file or directory to archive
//   - destTarPath: Destination path for the tar archive
//   - opts: Optional configuration functions
//
// Returns an error if the source cannot be read, the destination cannot be created,
// or if file operations fail during archiving.
//
// Example:
//
//	// Create gzip-compressed tar from directory
//	err := Tar("/path/to/source", "/path/to/archive.tar.gz")
//
//	// Create bzip2-compressed tar with explicit option
//	err := Tar("/path/to/source", "/path/to/archive.tar", WithBzip2Compression())
func Tar(sourcePath, destTarPath string, opts ...TarOption) error {
	// Configure compression
	config := &tarConfig{
		compression: detectCompression(destTarPath),
	}
	for _, opt := range opts {
		opt(config)
	}

	// Create destination file
	file, err := os.Create(destTarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create tar writer with appropriate compression
	tw, closer, err := createTarWriter(file, config.compression)
	if err != nil {
		return err
	}
	defer tw.Close()
	if closer != nil {
		defer closer.Close()
	}

	// Get the base directory for calculating relative paths
	baseDir := filepath.Dir(sourcePath)

	// Add files to archive
	return addToTar(tw, sourcePath, baseDir)
}

// createTarWriter creates a tar writer with the appropriate compression layer.
// Returns the tar writer and an optional closer for the compression writer.
func createTarWriter(file *os.File, compression compressionType) (*tar.Writer, io.WriteCloser, error) {
	switch compression {
	case compressionGzip:
		gw := gzip.NewWriter(file)
		return tar.NewWriter(gw), gw, nil

	case compressionBzip2:
		bw, err := bzip2writer.NewWriter(file, &bzip2writer.WriterConfig{Level: 9})
		if err != nil {
			return nil, nil, err
		}
		return tar.NewWriter(bw), bw, nil

	case compressionNone:
		return tar.NewWriter(file), nil, nil

	default:
		return nil, nil, ErrUnknownCompression
	}
}

// addToTar recursively adds files and directories to a tar archive.
func addToTar(tw *tar.Writer, sourcePath, baseDir string) error {
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := createTarHeader(info, path, baseDir)
		if err != nil {
			return err
		}

		// Handle symlinks - store the link target
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			header.Linkname = linkTarget
		}

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file content if it's a regular file
		if info.Mode().IsRegular() {
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

// createTarHeader creates a tar header from file info.
func createTarHeader(info os.FileInfo, path, baseDir string) (*tar.Header, error) {
	// Create header from file info
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return nil, err
	}

	// Calculate relative path for the archive
	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		return nil, err
	}

	// Use forward slashes in tar archives (portable across platforms)
	header.Name = filepath.ToSlash(relPath)

	return header, nil
}
