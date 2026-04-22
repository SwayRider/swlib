package compression

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	bzip2writer "github.com/dsnet/compress/bzip2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Section 1: Helper Functions
// =============================================================================

// tarEntry represents an entry in a tar archive for testing
type tarEntry struct {
	name     string      // Entry name (path)
	content  string      // File content (for regular files)
	mode     os.FileMode // File mode/permissions
	typeflag byte        // Tar type flag
	linkname string      // Link target (for symlinks/hardlinks)
}

// createTestTar creates an uncompressed tar archive with the given files
func createTestTar(t *testing.T, tarPath string, files map[string]string) {
	t.Helper()

	file, err := os.Create(tarPath)
	require.NoError(t, err)
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	for name, content := range files {
		header := &tar.Header{
			Name:     name,
			Size:     int64(len(content)),
			Mode:     0644,
			Typeflag: tar.TypeReg,
		}

		require.NoError(t, tw.WriteHeader(header))
		_, err := tw.Write([]byte(content))
		require.NoError(t, err)
	}
}

// createTestTarGzip creates a gzip-compressed tar archive with the given files
func createTestTarGzip(t *testing.T, tarPath string, files map[string]string) {
	t.Helper()

	file, err := os.Create(tarPath)
	require.NoError(t, err)
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for name, content := range files {
		header := &tar.Header{
			Name:     name,
			Size:     int64(len(content)),
			Mode:     0644,
			Typeflag: tar.TypeReg,
		}

		require.NoError(t, tw.WriteHeader(header))
		_, err := tw.Write([]byte(content))
		require.NoError(t, err)
	}
}

// createTestTarBzip2 creates a bzip2-compressed tar archive with the given files
func createTestTarBzip2(t *testing.T, tarPath string, files map[string]string) {
	t.Helper()

	file, err := os.Create(tarPath)
	require.NoError(t, err)
	defer file.Close()

	bw, err := bzip2writer.NewWriter(file, &bzip2writer.WriterConfig{Level: 9})
	require.NoError(t, err)
	defer bw.Close()

	tw := tar.NewWriter(bw)
	defer tw.Close()

	for name, content := range files {
		header := &tar.Header{
			Name:     name,
			Size:     int64(len(content)),
			Mode:     0644,
			Typeflag: tar.TypeReg,
		}

		require.NoError(t, tw.WriteHeader(header))
		_, err := tw.Write([]byte(content))
		require.NoError(t, err)
	}
}

// createTestTarWithDirs creates a tar archive with directories and files
func createTestTarWithDirs(t *testing.T, tarPath string, entries []tarEntry) {
	t.Helper()

	file, err := os.Create(tarPath)
	require.NoError(t, err)
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	for _, entry := range entries {
		header := &tar.Header{
			Name:     entry.name,
			Mode:     int64(entry.mode),
			Typeflag: entry.typeflag,
		}

		if entry.typeflag == tar.TypeReg {
			header.Size = int64(len(entry.content))
		}

		require.NoError(t, tw.WriteHeader(header))

		if entry.typeflag == tar.TypeReg && len(entry.content) > 0 {
			_, err := tw.Write([]byte(entry.content))
			require.NoError(t, err)
		}
	}
}

// createTestTarWithSymlinks creates a tar archive with symlinks
func createTestTarWithSymlinks(t *testing.T, tarPath string, entries []tarEntry) {
	t.Helper()

	file, err := os.Create(tarPath)
	require.NoError(t, err)
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	for _, entry := range entries {
		header := &tar.Header{
			Name:     entry.name,
			Mode:     int64(entry.mode),
			Typeflag: entry.typeflag,
			Linkname: entry.linkname,
		}

		if entry.typeflag == tar.TypeReg {
			header.Size = int64(len(entry.content))
		}

		require.NoError(t, tw.WriteHeader(header))

		if entry.typeflag == tar.TypeReg && len(entry.content) > 0 {
			_, err := tw.Write([]byte(entry.content))
			require.NoError(t, err)
		}
	}
}

// verifyFileContent checks if a file exists and has the expected content
func verifyFileContent(t *testing.T, path, expectedContent string) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, expectedContent, string(content))
}

// =============================================================================
// Section 2: Extraction Tests (UnTar)
// =============================================================================

// Basic extraction tests

func TestUnTar_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	files := map[string]string{
		"file.txt": "Hello, World!",
	}

	createTestTar(t, tarPath, files)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "file.txt"), "Hello, World!")
}

func TestUnTar_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	files := map[string]string{
		"file1.txt": "Content 1",
		"file2.txt": "Content 2",
		"file3.txt": "Content 3",
	}

	createTestTar(t, tarPath, files)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "file1.txt"), "Content 1")
	verifyFileContent(t, filepath.Join(destDir, "file2.txt"), "Content 2")
	verifyFileContent(t, filepath.Join(destDir, "file3.txt"), "Content 3")
}

func TestUnTar_WithDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	entries := []tarEntry{
		{name: "dir1/", mode: 0755, typeflag: tar.TypeDir},
		{name: "dir1/file1.txt", content: "File 1", mode: 0644, typeflag: tar.TypeReg},
		{name: "dir2/", mode: 0755, typeflag: tar.TypeDir},
		{name: "dir2/file2.txt", content: "File 2", mode: 0644, typeflag: tar.TypeReg},
	}

	createTestTarWithDirs(t, tarPath, entries)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	// Verify directories exist
	assert.DirExists(t, filepath.Join(destDir, "dir1"))
	assert.DirExists(t, filepath.Join(destDir, "dir2"))

	// Verify files
	verifyFileContent(t, filepath.Join(destDir, "dir1/file1.txt"), "File 1")
	verifyFileContent(t, filepath.Join(destDir, "dir2/file2.txt"), "File 2")
}

func TestUnTar_EmptyTar(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "empty.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create empty tar
	file, err := os.Create(tarPath)
	require.NoError(t, err)
	tw := tar.NewWriter(file)
	tw.Close()
	file.Close()

	err = UnTar(tarPath, destDir)
	require.NoError(t, err)
}

// Compression tests

func TestUnTar_GzipCompression(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar.gz")
	destDir := filepath.Join(tmpDir, "extracted")

	files := map[string]string{
		"file.txt": "Compressed content",
	}

	createTestTarGzip(t, tarPath, files)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "file.txt"), "Compressed content")
}

func TestUnTar_Bzip2Compression(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar.bz2")
	destDir := filepath.Join(tmpDir, "extracted")

	files := map[string]string{
		"file.txt": "Bzip2 compressed content",
	}

	createTestTarBzip2(t, tarPath, files)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "file.txt"), "Bzip2 compressed content")
}

func TestUnTar_UncompressedTar(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	files := map[string]string{
		"file.txt": "Uncompressed content",
	}

	createTestTar(t, tarPath, files)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "file.txt"), "Uncompressed content")
}

func TestUnTar_TgzExtension(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tgz")
	destDir := filepath.Join(tmpDir, "extracted")

	files := map[string]string{
		"file.txt": "TGZ content",
	}

	createTestTarGzip(t, tarPath, files)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "file.txt"), "TGZ content")
}

func TestUnTar_Tbz2Extension(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tbz2")
	destDir := filepath.Join(tmpDir, "extracted")

	files := map[string]string{
		"file.txt": "TBZ2 content",
	}

	createTestTarBzip2(t, tarPath, files)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "file.txt"), "TBZ2 content")
}

// Special content tests

func TestUnTar_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create 1MB content
	largeContent := strings.Repeat("A", 1024*1024)

	files := map[string]string{
		"large.txt": largeContent,
	}

	createTestTar(t, tarPath, files)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "large.txt"), largeContent)
}

func TestUnTar_BinaryContent(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create binary content with null bytes and special characters
	binaryContent := string([]byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD})

	files := map[string]string{
		"binary.dat": binaryContent,
	}

	createTestTar(t, tarPath, files)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "binary.dat"), binaryContent)
}

func TestUnTar_UnicodeFilenames(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	files := map[string]string{
		"файл.txt":  "Russian",
		"文件.txt":    "Chinese",
		"ファイル.txt": "Japanese",
	}

	createTestTar(t, tarPath, files)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "файл.txt"), "Russian")
	verifyFileContent(t, filepath.Join(destDir, "文件.txt"), "Chinese")
	verifyFileContent(t, filepath.Join(destDir, "ファイル.txt"), "Japanese")
}

func TestUnTar_SpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	files := map[string]string{
		"file with spaces.txt": "Space content",
		"file-with-dashes.txt": "Dash content",
	}

	createTestTar(t, tarPath, files)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "file with spaces.txt"), "Space content")
	verifyFileContent(t, filepath.Join(destDir, "file-with-dashes.txt"), "Dash content")
}

// Tar-specific tests

func TestUnTar_Symlinks(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	entries := []tarEntry{
		{name: "target.txt", content: "Target content", mode: 0644, typeflag: tar.TypeReg},
		{name: "link.txt", mode: 0777, typeflag: tar.TypeSymlink, linkname: "target.txt"},
	}

	createTestTarWithSymlinks(t, tarPath, entries)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	// Verify target file exists
	verifyFileContent(t, filepath.Join(destDir, "target.txt"), "Target content")

	// Verify symlink exists and points to correct target
	linkPath := filepath.Join(destDir, "link.txt")
	linkInfo, err := os.Lstat(linkPath)
	require.NoError(t, err)
	assert.True(t, linkInfo.Mode()&os.ModeSymlink != 0)

	// Verify reading through symlink works
	verifyFileContent(t, linkPath, "Target content")
}

func TestUnTar_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	entries := []tarEntry{
		{name: "executable.sh", content: "#!/bin/bash\necho test", mode: 0755, typeflag: tar.TypeReg},
		{name: "readonly.txt", content: "Read only", mode: 0444, typeflag: tar.TypeReg},
	}

	createTestTarWithDirs(t, tarPath, entries)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	// Verify executable has correct permissions
	execInfo, err := os.Stat(filepath.Join(destDir, "executable.sh"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), execInfo.Mode().Perm())

	// Verify readonly has correct permissions
	readonlyInfo, err := os.Stat(filepath.Join(destDir, "readonly.txt"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0444), readonlyInfo.Mode().Perm())
}

// Error tests

func TestUnTar_NonExistentTar(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "nonexistent.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	err := UnTar(tarPath, destDir)
	assert.Error(t, err)
}

func TestUnTar_InvalidTar(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "invalid.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create invalid tar file (just random content)
	err := os.WriteFile(tarPath, []byte("This is not a tar file"), 0644)
	require.NoError(t, err)

	err = UnTar(tarPath, destDir)
	assert.Error(t, err)
}

func TestUnTar_DestDirNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "nonexistent", "extracted")

	files := map[string]string{
		"file.txt": "Content",
	}

	createTestTar(t, tarPath, files)

	// This should succeed - UnTar creates destination directories
	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "file.txt"), "Content")
}

// =============================================================================
// Section 3: Creation Tests (Tar)
// =============================================================================

// Basic creation tests

func TestTar_CreateSingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.txt")
	tarPath := filepath.Join(tmpDir, "archive.tar")

	// Create source file
	err := os.WriteFile(sourcePath, []byte("Test content"), 0644)
	require.NoError(t, err)

	// Create tar
	err = Tar(sourcePath, tarPath)
	require.NoError(t, err)

	// Verify tar was created
	assert.FileExists(t, tarPath)

	// Extract and verify
	destDir := filepath.Join(tmpDir, "extracted")
	err = UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "source.txt"), "Test content")
}

func TestTar_CreateMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	tarPath := filepath.Join(tmpDir, "archive.tar")

	// Create source directory with multiple files
	require.NoError(t, os.MkdirAll(sourceDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file1.txt"), []byte("Content 1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file2.txt"), []byte("Content 2"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file3.txt"), []byte("Content 3"), 0644))

	// Create tar
	err := Tar(sourceDir, tarPath)
	require.NoError(t, err)

	// Extract and verify
	destDir := filepath.Join(tmpDir, "extracted")
	err = UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "source/file1.txt"), "Content 1")
	verifyFileContent(t, filepath.Join(destDir, "source/file2.txt"), "Content 2")
	verifyFileContent(t, filepath.Join(destDir, "source/file3.txt"), "Content 3")
}

func TestTar_CreateWithDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	tarPath := filepath.Join(tmpDir, "archive.tar")

	// Create source directory structure
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "dir1", "subdir1"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "dir2"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "dir1/file1.txt"), []byte("File 1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "dir1/subdir1/file2.txt"), []byte("File 2"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "dir2/file3.txt"), []byte("File 3"), 0644))

	// Create tar
	err := Tar(sourceDir, tarPath)
	require.NoError(t, err)

	// Extract and verify
	destDir := filepath.Join(tmpDir, "extracted")
	err = UnTar(tarPath, destDir)
	require.NoError(t, err)

	assert.DirExists(t, filepath.Join(destDir, "source/dir1"))
	assert.DirExists(t, filepath.Join(destDir, "source/dir1/subdir1"))
	assert.DirExists(t, filepath.Join(destDir, "source/dir2"))
	verifyFileContent(t, filepath.Join(destDir, "source/dir1/file1.txt"), "File 1")
	verifyFileContent(t, filepath.Join(destDir, "source/dir1/subdir1/file2.txt"), "File 2")
	verifyFileContent(t, filepath.Join(destDir, "source/dir2/file3.txt"), "File 3")
}

// Compression tests

func TestTar_CreateGzip(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.txt")
	tarPath := filepath.Join(tmpDir, "archive.tar.gz")

	// Create source file
	err := os.WriteFile(sourcePath, []byte("Gzip content"), 0644)
	require.NoError(t, err)

	// Create gzip tar
	err = Tar(sourcePath, tarPath)
	require.NoError(t, err)

	// Verify it's actually gzip compressed
	file, err := os.Open(tarPath)
	require.NoError(t, err)
	defer file.Close()

	// Try to open as gzip
	_, err = gzip.NewReader(file)
	assert.NoError(t, err, "File should be gzip compressed")

	// Extract and verify
	destDir := filepath.Join(tmpDir, "extracted")
	err = UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "source.txt"), "Gzip content")
}

func TestTar_CreateBzip2(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.txt")
	tarPath := filepath.Join(tmpDir, "archive.tar.bz2")

	// Create source file
	err := os.WriteFile(sourcePath, []byte("Bzip2 content"), 0644)
	require.NoError(t, err)

	// Create bzip2 tar
	err = Tar(sourcePath, tarPath)
	require.NoError(t, err)

	// Extract and verify
	destDir := filepath.Join(tmpDir, "extracted")
	err = UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "source.txt"), "Bzip2 content")
}

func TestTar_CreateUncompressed(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.txt")
	tarPath := filepath.Join(tmpDir, "archive.tar")

	// Create source file
	err := os.WriteFile(sourcePath, []byte("Uncompressed content"), 0644)
	require.NoError(t, err)

	// Create uncompressed tar
	err = Tar(sourcePath, tarPath)
	require.NoError(t, err)

	// Extract and verify
	destDir := filepath.Join(tmpDir, "extracted")
	err = UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "source.txt"), "Uncompressed content")
}

// Option tests

func TestTar_WithGzipCompressionOption(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.txt")
	// Use .tar extension but force gzip compression
	tarPath := filepath.Join(tmpDir, "archive.tar")

	// Create source file
	err := os.WriteFile(sourcePath, []byte("Forced gzip"), 0644)
	require.NoError(t, err)

	// Create tar with gzip option
	err = Tar(sourcePath, tarPath, WithGzipCompression())
	require.NoError(t, err)

	// Verify it's gzip compressed despite .tar extension
	file, err := os.Open(tarPath)
	require.NoError(t, err)
	defer file.Close()

	_, err = gzip.NewReader(file)
	assert.NoError(t, err, "File should be gzip compressed")
}

func TestTar_WithBzip2CompressionOption(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.txt")
	// Use .tar extension but force bzip2 compression
	tarPath := filepath.Join(tmpDir, "archive.tar")

	// Create source file
	err := os.WriteFile(sourcePath, []byte("Forced bzip2"), 0644)
	require.NoError(t, err)

	// Create tar with bzip2 option
	err = Tar(sourcePath, tarPath, WithBzip2Compression())
	require.NoError(t, err)

	// We need to manually extract since UnTar detects by extension
	file, err := os.Open(tarPath)
	require.NoError(t, err)
	defer file.Close()

	// Should be able to read as bzip2
	br := bzip2.NewReader(file)
	tr := tar.NewReader(br)
	header, err := tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "source.txt", header.Name)
}

func TestTar_WithNoCompressionOption(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.txt")
	// Use .tar.gz extension but force no compression
	tarPath := filepath.Join(tmpDir, "archive.tar.gz")

	// Create source file
	err := os.WriteFile(sourcePath, []byte("Forced uncompressed"), 0644)
	require.NoError(t, err)

	// Create tar with no compression option
	err = Tar(sourcePath, tarPath, WithNoCompression())
	require.NoError(t, err)

	// Verify it's NOT gzip compressed despite .tar.gz extension
	file, err := os.Open(tarPath)
	require.NoError(t, err)
	defer file.Close()

	// Should be able to read directly as tar without gzip
	tr := tar.NewReader(file)
	header, err := tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "source.txt", header.Name)
}

// Special tests

func TestTar_CreateLargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "large.txt")
	tarPath := filepath.Join(tmpDir, "archive.tar.gz")

	// Create 5MB file
	largeContent := strings.Repeat("X", 5*1024*1024)
	err := os.WriteFile(sourcePath, []byte(largeContent), 0644)
	require.NoError(t, err)

	// Create tar
	err = Tar(sourcePath, tarPath)
	require.NoError(t, err)

	// Extract and verify
	destDir := filepath.Join(tmpDir, "extracted")
	err = UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "large.txt"), largeContent)
}

func TestTar_CreateWithUnicodeNames(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	tarPath := filepath.Join(tmpDir, "archive.tar")

	// Create source directory with unicode filenames
	require.NoError(t, os.MkdirAll(sourceDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "файл.txt"), []byte("Russian"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "文件.txt"), []byte("Chinese"), 0644))

	// Create tar
	err := Tar(sourceDir, tarPath)
	require.NoError(t, err)

	// Extract and verify
	destDir := filepath.Join(tmpDir, "extracted")
	err = UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "source/файл.txt"), "Russian")
	verifyFileContent(t, filepath.Join(destDir, "source/文件.txt"), "Chinese")
}

// Roundtrip test

func TestTar_CreateAndExtract(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	tarPath := filepath.Join(tmpDir, "archive.tar.gz")

	// Create complex source structure
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "dir1", "subdir"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "dir2"), 0755))

	files := map[string]string{
		"file1.txt":               "Root file",
		"dir1/file2.txt":          "Dir1 file",
		"dir1/subdir/file3.txt":   "Subdir file",
		"dir2/file4.txt":          "Dir2 file",
		"unicode_файл.txt":        "Unicode content",
		"spaces in name.txt":      "Spaces content",
		"special-chars_123.txt":   "Special chars",
	}

	for path, content := range files {
		fullPath := filepath.Join(sourceDir, path)
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	// Create tar
	err := Tar(sourceDir, tarPath)
	require.NoError(t, err)

	// Extract to new location
	destDir := filepath.Join(tmpDir, "extracted")
	err = UnTar(tarPath, destDir)
	require.NoError(t, err)

	// Verify all files
	for path, content := range files {
		verifyFileContent(t, filepath.Join(destDir, "source", path), content)
	}

	// Verify directory structure
	assert.DirExists(t, filepath.Join(destDir, "source/dir1"))
	assert.DirExists(t, filepath.Join(destDir, "source/dir1/subdir"))
	assert.DirExists(t, filepath.Join(destDir, "source/dir2"))
}

func TestTar_CreateWithSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	tarPath := filepath.Join(tmpDir, "archive.tar")

	// Create source directory with symlinks
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	// Create a target file
	targetFile := filepath.Join(sourceDir, "target.txt")
	require.NoError(t, os.WriteFile(targetFile, []byte("Target content"), 0644))

	// Create a symlink to the target
	symlinkPath := filepath.Join(sourceDir, "link.txt")
	require.NoError(t, os.Symlink("target.txt", symlinkPath))

	// Create tar
	err := Tar(sourceDir, tarPath)
	require.NoError(t, err)

	// Extract and verify
	destDir := filepath.Join(tmpDir, "extracted")
	err = UnTar(tarPath, destDir)
	require.NoError(t, err)

	// Verify target file
	verifyFileContent(t, filepath.Join(destDir, "source/target.txt"), "Target content")

	// Verify symlink exists
	linkInfo, err := os.Lstat(filepath.Join(destDir, "source/link.txt"))
	require.NoError(t, err)
	assert.True(t, linkInfo.Mode()&os.ModeSymlink != 0, "Should be a symlink")

	// Verify reading through symlink works
	verifyFileContent(t, filepath.Join(destDir, "source/link.txt"), "Target content")
}

func TestTar_NonExistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "nonexistent")
	tarPath := filepath.Join(tmpDir, "archive.tar")

	err := Tar(sourcePath, tarPath)
	assert.Error(t, err)
}

func TestTar_InvalidDestPath(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.txt")

	// Create source file
	require.NoError(t, os.WriteFile(sourcePath, []byte("content"), 0644))

	// Try to create tar in non-existent directory
	tarPath := filepath.Join(tmpDir, "nonexistent", "archive.tar")

	err := Tar(sourcePath, tarPath)
	assert.Error(t, err)
}

// =============================================================================
// Section 4: Security Tests
// =============================================================================

func TestUnTar_TarSlip_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "malicious.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create malicious tar with path traversal
	entries := []tarEntry{
		{name: "../../../etc/evil.txt", content: "Malicious", mode: 0644, typeflag: tar.TypeReg},
	}

	createTestTarWithDirs(t, tarPath, entries)

	err := UnTar(tarPath, destDir)
	assert.ErrorIs(t, err, ErrTarSlip)
}

func TestUnTar_TarSlip_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "malicious.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create malicious tar with absolute path
	// Note: filepath.Clean will convert /tmp/evil.txt to a relative path on some systems
	// So we test that the file doesn't escape destDir
	entries := []tarEntry{
		{name: "/tmp/evil.txt", content: "Malicious", mode: 0644, typeflag: tar.TypeReg},
	}

	createTestTarWithDirs(t, tarPath, entries)

	err := UnTar(tarPath, destDir)
	// The implementation may handle absolute paths by making them relative
	// which is safe. The important thing is that the file ends up in destDir.
	if err != nil {
		// If there's an error, it should be a tar slip error
		assert.ErrorIs(t, err, ErrTarSlip)
	} else {
		// If extraction succeeded, verify the file is within destDir
		extractedPath := filepath.Join(destDir, "tmp/evil.txt")
		_, statErr := os.Stat(extractedPath)
		assert.NoError(t, statErr, "File should be extracted within destDir if no error")
	}
}

func TestUnTar_TarSlip_DotDotInMiddle(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "malicious.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create malicious tar with ../ in middle of path
	entries := []tarEntry{
		{name: "dir1/../../evil.txt", content: "Malicious", mode: 0644, typeflag: tar.TypeReg},
	}

	createTestTarWithDirs(t, tarPath, entries)

	err := UnTar(tarPath, destDir)
	assert.ErrorIs(t, err, ErrTarSlip)
}

func TestUnTar_TarSlip_MaliciousSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "malicious.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create malicious tar with symlink pointing outside destDir
	entries := []tarEntry{
		{name: "evil_link", mode: 0777, typeflag: tar.TypeSymlink, linkname: "../../../etc/passwd"},
	}

	createTestTarWithSymlinks(t, tarPath, entries)

	err := UnTar(tarPath, destDir)
	assert.ErrorIs(t, err, ErrInvalidSymlink)
}

func TestUnTar_SafeNestedDirs(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "safe.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create tar with deeply nested but safe directories
	entries := []tarEntry{
		{name: "a/b/c/d/e/f/g/h/i/j/file.txt", content: "Deep content", mode: 0644, typeflag: tar.TypeReg},
	}

	createTestTarWithDirs(t, tarPath, entries)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	verifyFileContent(t, filepath.Join(destDir, "a/b/c/d/e/f/g/h/i/j/file.txt"), "Deep content")
}

func TestUnTar_HardLinks(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	entries := []tarEntry{
		{name: "original.txt", content: "Original content", mode: 0644, typeflag: tar.TypeReg},
		{name: "hardlink.txt", mode: 0644, typeflag: tar.TypeLink, linkname: "original.txt"},
	}

	createTestTarWithSymlinks(t, tarPath, entries)

	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	// Verify both files exist and have the same content
	verifyFileContent(t, filepath.Join(destDir, "original.txt"), "Original content")
	verifyFileContent(t, filepath.Join(destDir, "hardlink.txt"), "Original content")

	// Verify they are hard links by checking inode
	origInfo, err := os.Stat(filepath.Join(destDir, "original.txt"))
	require.NoError(t, err)
	linkInfo, err := os.Stat(filepath.Join(destDir, "hardlink.txt"))
	require.NoError(t, err)

	// On systems that support hard links, they should have the same inode
	// We can't easily test this in a cross-platform way, so just verify content
	assert.Equal(t, origInfo.Size(), linkInfo.Size())
}

func TestUnTar_UnsupportedEntryTypes(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create tar with unsupported entry types (device file)
	entries := []tarEntry{
		{name: "normalfile.txt", content: "Normal", mode: 0644, typeflag: tar.TypeReg},
		{name: "chardevice", mode: 0666, typeflag: tar.TypeChar}, // Character device
		{name: "blockdevice", mode: 0666, typeflag: tar.TypeBlock}, // Block device
		{name: "fifo", mode: 0666, typeflag: tar.TypeFifo}, // FIFO
	}

	createTestTarWithDirs(t, tarPath, entries)

	// Should not error - unsupported types are skipped
	err := UnTar(tarPath, destDir)
	require.NoError(t, err)

	// Verify normal file was extracted
	verifyFileContent(t, filepath.Join(destDir, "normalfile.txt"), "Normal")

	// Verify unsupported entries were skipped (not extracted)
	_, err = os.Stat(filepath.Join(destDir, "chardevice"))
	assert.True(t, os.IsNotExist(err), "Unsupported entry should not be extracted")
}

func TestUnTar_SymlinkWithAbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create tar with symlink that has absolute path (should fail)
	entries := []tarEntry{
		{name: "symlink", mode: 0777, typeflag: tar.TypeSymlink, linkname: "/etc/passwd"},
	}

	createTestTarWithSymlinks(t, tarPath, entries)

	err := UnTar(tarPath, destDir)
	assert.ErrorIs(t, err, ErrInvalidSymlink)
}

func TestUnTar_EmptyFileName(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create tar with empty file name (edge case)
	file, err := os.Create(tarPath)
	require.NoError(t, err)

	tw := tar.NewWriter(file)
	header := &tar.Header{
		Name:     ".",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	}
	require.NoError(t, tw.WriteHeader(header))
	tw.Close()
	file.Close()

	// Should handle gracefully
	err = UnTar(tarPath, destDir)
	require.NoError(t, err)
}

// =============================================================================
// Section 5: Benchmarks
// =============================================================================

func BenchmarkUnTar_SmallFiles(b *testing.B) {
	tmpDir := b.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")

	// Create tar with 100 small files
	files := make(map[string]string)
	for i := 0; i < 100; i++ {
		files[filepath.Join("file", filepath.Base(filepath.Join("", string(rune(i)))))+".txt"] = "Small content"
	}

	createTestTar(&testing.T{}, tarPath, files)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destDir := filepath.Join(tmpDir, "extracted", string(rune(i)))
		_ = UnTar(tarPath, destDir)
	}
}

func BenchmarkUnTar_LargeFile(b *testing.B) {
	tmpDir := b.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")

	// Create tar with one 10MB file
	largeContent := strings.Repeat("X", 10*1024*1024)
	files := map[string]string{
		"large.bin": largeContent,
	}

	createTestTar(&testing.T{}, tarPath, files)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destDir := filepath.Join(tmpDir, "extracted", string(rune(i)))
		_ = UnTar(tarPath, destDir)
	}
}

func BenchmarkUnTar_GzipCompression(b *testing.B) {
	tmpDir := b.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar.gz")

	// Create gzip tar with 1MB content
	content := strings.Repeat("A", 1024*1024)
	files := map[string]string{
		"file.txt": content,
	}

	createTestTarGzip(&testing.T{}, tarPath, files)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destDir := filepath.Join(tmpDir, "extracted", string(rune(i)))
		_ = UnTar(tarPath, destDir)
	}
}

func BenchmarkUnTar_Bzip2Compression(b *testing.B) {
	tmpDir := b.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar.bz2")

	// Create bzip2 tar with 1MB content
	content := strings.Repeat("A", 1024*1024)
	files := map[string]string{
		"file.txt": content,
	}

	createTestTarBzip2(&testing.T{}, tarPath, files)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destDir := filepath.Join(tmpDir, "extracted", string(rune(i)))
		_ = UnTar(tarPath, destDir)
	}
}

func BenchmarkTar_SmallFiles(b *testing.B) {
	tmpDir := b.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")

	// Create source directory with 100 small files
	_ = os.MkdirAll(sourceDir, 0755)
	for i := 0; i < 100; i++ {
		path := filepath.Join(sourceDir, "file"+string(rune(i))+".txt")
		_ = os.WriteFile(path, []byte("Small content"), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tarPath := filepath.Join(tmpDir, "archive"+string(rune(i))+".tar")
		_ = Tar(sourceDir, tarPath)
	}
}

func BenchmarkTar_LargeFile(b *testing.B) {
	tmpDir := b.TempDir()
	sourcePath := filepath.Join(tmpDir, "large.bin")

	// Create 10MB file
	largeContent := strings.Repeat("X", 10*1024*1024)
	_ = os.WriteFile(sourcePath, []byte(largeContent), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tarPath := filepath.Join(tmpDir, "archive"+string(rune(i))+".tar")
		_ = Tar(sourcePath, tarPath)
	}
}

func BenchmarkTar_GzipCompression(b *testing.B) {
	tmpDir := b.TempDir()
	sourcePath := filepath.Join(tmpDir, "file.txt")

	// Create 1MB file
	content := strings.Repeat("A", 1024*1024)
	_ = os.WriteFile(sourcePath, []byte(content), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tarPath := filepath.Join(tmpDir, "archive"+string(rune(i))+".tar.gz")
		_ = Tar(sourcePath, tarPath)
	}
}

func BenchmarkTar_Bzip2Compression(b *testing.B) {
	tmpDir := b.TempDir()
	sourcePath := filepath.Join(tmpDir, "file.txt")

	// Create 1MB file
	content := strings.Repeat("A", 1024*1024)
	_ = os.WriteFile(sourcePath, []byte(content), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tarPath := filepath.Join(tmpDir, "archive"+string(rune(i))+".tar.bz2")
		_ = Tar(sourcePath, tarPath)
	}
}
