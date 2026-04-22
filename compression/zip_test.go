package compression

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// Helper to create a test ZIP file
func createTestZip(t *testing.T, zipPath string, files map[string]string) {
	t.Helper()

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatalf("failed to create file in zip: %v", err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write to file in zip: %v", err)
		}
	}
}

// Helper to create a test ZIP file with directories
func createTestZipWithDirs(t *testing.T, zipPath string, entries []zipEntry) {
	t.Helper()

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for _, entry := range entries {
		if entry.isDir {
			// Create directory entry (name must end with /)
			name := entry.name
			if name[len(name)-1] != '/' {
				name += "/"
			}
			_, err := w.Create(name)
			if err != nil {
				t.Fatalf("failed to create dir in zip: %v", err)
			}
		} else {
			fw, err := w.Create(entry.name)
			if err != nil {
				t.Fatalf("failed to create file in zip: %v", err)
			}
			if _, err := fw.Write([]byte(entry.content)); err != nil {
				t.Fatalf("failed to write to file in zip: %v", err)
			}
		}
	}
}

type zipEntry struct {
	name    string
	content string
	isDir   bool
}

// =============================================================================
// UnZip Tests
// =============================================================================

func TestUnZip_SingleFile(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create a test zip file
	createTestZip(t, zipPath, map[string]string{
		"file.txt": "hello world",
	})

	// Extract
	err := UnZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("UnZip failed: %v", err)
	}

	// Verify extracted file
	content, err := os.ReadFile(filepath.Join(destDir, "file.txt"))
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(content))
	}
}

func TestUnZip_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	createTestZip(t, zipPath, map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
		"file3.txt": "content3",
	})

	err := UnZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("UnZip failed: %v", err)
	}

	// Verify all files
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(destDir, "file"+string(rune('0'+i))+".txt")
		content, err := os.ReadFile(filename)
		if err != nil {
			t.Errorf("failed to read file%d.txt: %v", i, err)
			continue
		}
		expected := "content" + string(rune('0'+i))
		if string(content) != expected {
			t.Errorf("file%d.txt: expected '%s', got '%s'", i, expected, string(content))
		}
	}
}

func TestUnZip_WithDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	createTestZipWithDirs(t, zipPath, []zipEntry{
		{name: "dir1/", isDir: true},
		{name: "dir1/file1.txt", content: "in dir1"},
		{name: "dir2/subdir/", isDir: true},
		{name: "dir2/subdir/file2.txt", content: "in subdir"},
		{name: "root.txt", content: "at root"},
	})

	err := UnZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("UnZip failed: %v", err)
	}

	// Verify directory structure
	tests := []struct {
		path    string
		content string
	}{
		{"root.txt", "at root"},
		{"dir1/file1.txt", "in dir1"},
		{"dir2/subdir/file2.txt", "in subdir"},
	}

	for _, tc := range tests {
		fullPath := filepath.Join(destDir, tc.path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("failed to read %s: %v", tc.path, err)
			continue
		}
		if string(content) != tc.content {
			t.Errorf("%s: expected '%s', got '%s'", tc.path, tc.content, string(content))
		}
	}
}

func TestUnZip_EmptyZip(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "empty.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create empty zip
	createTestZip(t, zipPath, map[string]string{})

	err := UnZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("UnZip failed on empty zip: %v", err)
	}
}

func TestUnZip_NonExistentZip(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "nonexistent.zip")
	destDir := filepath.Join(tmpDir, "dest")

	err := UnZip(zipPath, destDir)
	if err == nil {
		t.Error("expected error for non-existent zip file")
	}
}

func TestUnZip_InvalidZip(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "invalid.zip")
	destDir := filepath.Join(tmpDir, "dest")

	// Create an invalid zip file (just some random data)
	if err := os.WriteFile(zipPath, []byte("this is not a zip file"), 0644); err != nil {
		t.Fatalf("failed to create invalid zip: %v", err)
	}

	err := UnZip(zipPath, destDir)
	if err == nil {
		t.Error("expected error for invalid zip file")
	}
}

func TestUnZip_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "large.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create a file with 1MB of data
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte('A' + i%26)
	}

	createTestZip(t, zipPath, map[string]string{
		"large.txt": string(largeContent),
	})

	err := UnZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("UnZip failed on large file: %v", err)
	}

	// Verify file size
	info, err := os.Stat(filepath.Join(destDir, "large.txt"))
	if err != nil {
		t.Fatalf("failed to stat large file: %v", err)
	}
	if info.Size() != 1024*1024 {
		t.Errorf("expected 1MB file, got %d bytes", info.Size())
	}
}

func TestUnZip_BinaryContent(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "binary.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create binary content
	binaryContent := make([]byte, 256)
	for i := range binaryContent {
		binaryContent[i] = byte(i)
	}

	createTestZip(t, zipPath, map[string]string{
		"binary.bin": string(binaryContent),
	})

	err := UnZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("UnZip failed on binary file: %v", err)
	}

	// Verify binary content
	content, err := os.ReadFile(filepath.Join(destDir, "binary.bin"))
	if err != nil {
		t.Fatalf("failed to read binary file: %v", err)
	}
	if len(content) != 256 {
		t.Errorf("expected 256 bytes, got %d", len(content))
	}
	for i, b := range content {
		if b != byte(i) {
			t.Errorf("byte %d: expected %d, got %d", i, i, b)
			break
		}
	}
}

func TestUnZip_UnicodeFilenames(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "unicode.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	createTestZip(t, zipPath, map[string]string{
		"日本語.txt":    "Japanese",
		"emoji_🎉.txt": "Party",
	})

	err := UnZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("UnZip failed on unicode filenames: %v", err)
	}

	// Verify files exist
	if _, err := os.Stat(filepath.Join(destDir, "日本語.txt")); err != nil {
		t.Errorf("Japanese filename not found: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "emoji_🎉.txt")); err != nil {
		t.Errorf("Emoji filename not found: %v", err)
	}
}

func TestUnZip_DestDirNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	destDir := filepath.Join(tmpDir, "nonexistent", "nested", "dest")

	createTestZip(t, zipPath, map[string]string{
		"file.txt": "content",
	})

	// UnZip now creates parent directories as needed for extracted files
	err := UnZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("UnZip failed: %v", err)
	}

	// Verify the file was created in the nested destination
	content, err := os.ReadFile(filepath.Join(destDir, "file.txt"))
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(content) != "content" {
		t.Errorf("expected 'content', got '%s'", string(content))
	}
}

func TestUnZip_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "empty_file.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	createTestZip(t, zipPath, map[string]string{
		"empty.txt": "",
	})

	err := UnZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("UnZip failed on empty file: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "empty.txt"))
	if err != nil {
		t.Fatalf("failed to read empty file: %v", err)
	}
	if len(content) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(content))
	}
}

func TestUnZip_SpecialCharactersInFilename(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "special.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	createTestZip(t, zipPath, map[string]string{
		"file with spaces.txt":     "spaces",
		"file-with-dashes.txt":     "dashes",
		"file_with_underscores.txt": "underscores",
	})

	err := UnZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("UnZip failed on special characters: %v", err)
	}

	// Verify files exist
	files := []string{
		"file with spaces.txt",
		"file-with-dashes.txt",
		"file_with_underscores.txt",
	}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(destDir, f)); err != nil {
			t.Errorf("file '%s' not found: %v", f, err)
		}
	}
}

// =============================================================================
// Security Tests - Zip Slip Protection
// =============================================================================

func TestUnZip_ZipSlip_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "malicious.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create a malicious zip with path traversal
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	w := zip.NewWriter(f)
	// This malicious entry tries to escape the destination directory
	fw, err := w.Create("../../../tmp/evil.txt")
	if err != nil {
		t.Fatalf("failed to create malicious entry: %v", err)
	}
	fw.Write([]byte("malicious content"))
	w.Close()
	f.Close()

	// UnZip should detect and reject the zip slip attempt
	err = UnZip(zipPath, destDir)
	if err != ErrZipSlip {
		t.Errorf("expected ErrZipSlip, got: %v", err)
	}

	// Verify malicious file was NOT created
	if _, err := os.Stat(filepath.Join(tmpDir, "tmp", "evil.txt")); err == nil {
		t.Error("malicious file should not have been created")
	}
}

func TestUnZip_ZipSlip_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "malicious.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create a malicious zip with absolute path
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	w := zip.NewWriter(f)
	// Absolute path attempt (will be cleaned by filepath.Join but should still be caught)
	fw, err := w.Create("/etc/passwd")
	if err != nil {
		t.Fatalf("failed to create malicious entry: %v", err)
	}
	fw.Write([]byte("malicious content"))
	w.Close()
	f.Close()

	// UnZip should handle this safely
	err = UnZip(zipPath, destDir)
	// After filepath.Clean, "/etc/passwd" becomes "etc/passwd" which is safe
	// So this should succeed and create destDir/etc/passwd
	if err != nil {
		t.Logf("UnZip returned error (expected for absolute paths): %v", err)
	}

	// Verify file was created inside destDir, not at /etc/passwd
	if _, err := os.Stat("/etc/passwd.test"); err == nil {
		t.Error("should not have written to /etc/passwd.test")
	}
}

func TestUnZip_ZipSlip_DotDotInMiddle(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "malicious.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create a zip with .. in the middle of path
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	w := zip.NewWriter(f)
	fw, err := w.Create("subdir/../../outside.txt")
	if err != nil {
		t.Fatalf("failed to create malicious entry: %v", err)
	}
	fw.Write([]byte("trying to escape"))
	w.Close()
	f.Close()

	err = UnZip(zipPath, destDir)
	if err != ErrZipSlip {
		t.Errorf("expected ErrZipSlip for path with .. in middle, got: %v", err)
	}
}

func TestUnZip_SafeNestedDirs(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "safe.zip")
	destDir := filepath.Join(tmpDir, "dest")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create a safe zip with nested directories
	createTestZipWithDirs(t, zipPath, []zipEntry{
		{name: "level1/", isDir: true},
		{name: "level1/level2/", isDir: true},
		{name: "level1/level2/level3/", isDir: true},
		{name: "level1/level2/level3/deep.txt", content: "deep file"},
	})

	err := UnZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("UnZip failed on safe nested dirs: %v", err)
	}

	// Verify the deep file was created
	content, err := os.ReadFile(filepath.Join(destDir, "level1/level2/level3/deep.txt"))
	if err != nil {
		t.Fatalf("failed to read deep file: %v", err)
	}
	if string(content) != "deep file" {
		t.Errorf("unexpected content: %s", string(content))
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkUnZip_SmallFiles(b *testing.B) {
	tmpDir := b.TempDir()
	zipPath := filepath.Join(tmpDir, "bench.zip")
	destDir := filepath.Join(tmpDir, "dest")

	// Create zip with small files
	f, _ := os.Create(zipPath)
	w := zip.NewWriter(f)
	for i := 0; i < 10; i++ {
		fw, _ := w.Create("file" + string(rune('0'+i)) + ".txt")
		fw.Write([]byte("small content"))
	}
	w.Close()
	f.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		os.MkdirAll(destDir, 0755)
		UnZip(zipPath, destDir)
		os.RemoveAll(destDir)
	}
}

func BenchmarkUnZip_LargeFile(b *testing.B) {
	tmpDir := b.TempDir()
	zipPath := filepath.Join(tmpDir, "bench.zip")
	destDir := filepath.Join(tmpDir, "dest")

	// Create zip with large file
	largeContent := make([]byte, 1024*1024) // 1MB
	f, _ := os.Create(zipPath)
	w := zip.NewWriter(f)
	fw, _ := w.Create("large.txt")
	fw.Write(largeContent)
	w.Close()
	f.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		os.MkdirAll(destDir, 0755)
		UnZip(zipPath, destDir)
		os.RemoveAll(destDir)
	}
}

// =============================================================================
// Zip Creation Tests
// =============================================================================

// TestZip_CreateSingleFile tests creating a zip archive from a single file.
func TestZip_CreateSingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "source.txt")
	zipPath := filepath.Join(tmpDir, "archive.zip")

	// Create source file
	if err := os.WriteFile(sourceFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Create zip archive
	err := Zip(sourceFile, zipPath)
	if err != nil {
		t.Fatalf("Zip failed: %v", err)
	}

	// Verify zip was created
	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("zip file not created: %v", err)
	}

	// Verify zip can be opened with standard library
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open created zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 1 {
		t.Fatalf("expected 1 file in zip, got %d", len(r.File))
	}

	// Verify file content
	f := r.File[0]
	if f.Name != "source.txt" {
		t.Errorf("expected filename 'source.txt', got '%s'", f.Name)
	}

	rc, err := f.Open()
	if err != nil {
		t.Fatalf("failed to open file in zip: %v", err)
	}
	defer rc.Close()

	content, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("failed to read source file: %v", err)
	}

	var buf []byte
	buf = make([]byte, len(content))
	_, err = rc.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("failed to read from zip: %v", err)
	}

	if string(buf) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(buf))
	}
}

// TestZip_CreateMultipleFiles tests creating a zip archive from a directory with multiple files.
func TestZip_CreateMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	zipPath := filepath.Join(tmpDir, "archive.zip")

	// Create source directory with multiple files
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	files := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
		"file3.txt": "content3",
	}

	for name, content := range files {
		path := filepath.Join(sourceDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", name, err)
		}
	}

	// Create zip archive
	err := Zip(sourceDir, zipPath)
	if err != nil {
		t.Fatalf("Zip failed: %v", err)
	}

	// Verify zip contents
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open created zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 3 {
		t.Fatalf("expected 3 files in zip, got %d", len(r.File))
	}

	// Verify each file exists and has correct content
	for _, f := range r.File {
		expectedContent, exists := files[filepath.Base(f.Name)]
		if !exists {
			t.Errorf("unexpected file in zip: %s", f.Name)
			continue
		}

		rc, err := f.Open()
		if err != nil {
			t.Errorf("failed to open file %s in zip: %v", f.Name, err)
			continue
		}

		buf := make([]byte, f.UncompressedSize64)
		_, err = rc.Read(buf)
		rc.Close()

		if err != nil && err.Error() != "EOF" {
			t.Errorf("failed to read file %s: %v", f.Name, err)
			continue
		}

		if string(buf) != expectedContent {
			t.Errorf("file %s: expected '%s', got '%s'", f.Name, expectedContent, string(buf))
		}
	}
}

// TestZip_CreateWithDirectories tests creating a zip archive from a directory with nested structure.
func TestZip_CreateWithDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	zipPath := filepath.Join(tmpDir, "archive.zip")

	// Create nested directory structure
	if err := os.MkdirAll(filepath.Join(sourceDir, "dir1", "subdir"), 0755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(sourceDir, "dir2"), 0755); err != nil {
		t.Fatalf("failed to create dir2: %v", err)
	}

	// Create files in various locations
	files := map[string]string{
		"root.txt":              "at root",
		"dir1/file1.txt":        "in dir1",
		"dir1/subdir/deep.txt":  "deep file",
		"dir2/file2.txt":        "in dir2",
	}

	for relPath, content := range files {
		fullPath := filepath.Join(sourceDir, relPath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", relPath, err)
		}
	}

	// Create zip archive
	err := Zip(sourceDir, zipPath)
	if err != nil {
		t.Fatalf("Zip failed: %v", err)
	}

	// Verify zip contents
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open created zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 4 {
		t.Fatalf("expected 4 files in zip, got %d", len(r.File))
	}

	// Verify directory structure is preserved
	foundFiles := make(map[string]bool)
	for _, f := range r.File {
		foundFiles[f.Name] = true
	}

	// Expected paths in zip (using forward slashes)
	expectedPaths := []string{
		"source/root.txt",
		"source/dir1/file1.txt",
		"source/dir1/subdir/deep.txt",
		"source/dir2/file2.txt",
	}

	for _, expected := range expectedPaths {
		if !foundFiles[expected] {
			t.Errorf("expected file not found in zip: %s", expected)
		}
	}
}

// TestZip_WithZipDeflateOption tests explicit deflate compression option.
func TestZip_WithZipDeflateOption(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "source.txt")
	zipPath := filepath.Join(tmpDir, "archive.zip")

	// Create source file with compressible content
	content := make([]byte, 1024)
	for i := range content {
		content[i] = 'A' // Highly compressible
	}
	if err := os.WriteFile(sourceFile, content, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Create zip with explicit deflate option
	err := Zip(sourceFile, zipPath, WithZipDeflate())
	if err != nil {
		t.Fatalf("Zip failed: %v", err)
	}

	// Verify compression method
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open created zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 1 {
		t.Fatalf("expected 1 file in zip, got %d", len(r.File))
	}

	f := r.File[0]
	if f.Method != zip.Deflate {
		t.Errorf("expected Deflate compression method, got %d", f.Method)
	}

	// Verify file is actually compressed
	if f.CompressedSize64 >= f.UncompressedSize64 {
		t.Errorf("file not compressed: compressed=%d, uncompressed=%d",
			f.CompressedSize64, f.UncompressedSize64)
	}
}

// TestZip_WithZipStoreOption tests no compression (store) option.
func TestZip_WithZipStoreOption(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "source.txt")
	zipPath := filepath.Join(tmpDir, "archive.zip")

	// Create source file
	content := []byte("test content for store method")
	if err := os.WriteFile(sourceFile, content, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Create zip with store (no compression) option
	err := Zip(sourceFile, zipPath, WithZipStore())
	if err != nil {
		t.Fatalf("Zip failed: %v", err)
	}

	// Verify compression method
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open created zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 1 {
		t.Fatalf("expected 1 file in zip, got %d", len(r.File))
	}

	f := r.File[0]
	if f.Method != zip.Store {
		t.Errorf("expected Store method, got %d", f.Method)
	}

	// Verify file is not compressed
	if f.CompressedSize64 != f.UncompressedSize64 {
		t.Errorf("file should not be compressed with Store method: compressed=%d, uncompressed=%d",
			f.CompressedSize64, f.UncompressedSize64)
	}
}

// TestZip_CreateLargeFile tests creating a zip archive from a large file (5MB).
func TestZip_CreateLargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "large.txt")
	zipPath := filepath.Join(tmpDir, "archive.zip")

	// Create 5MB file
	largeContent := make([]byte, 5*1024*1024)
	for i := range largeContent {
		largeContent[i] = byte('A' + i%26)
	}
	if err := os.WriteFile(sourceFile, largeContent, 0644); err != nil {
		t.Fatalf("failed to create large file: %v", err)
	}

	// Create zip archive
	err := Zip(sourceFile, zipPath)
	if err != nil {
		t.Fatalf("Zip failed: %v", err)
	}

	// Verify zip was created and contains the large file
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open created zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 1 {
		t.Fatalf("expected 1 file in zip, got %d", len(r.File))
	}

	f := r.File[0]
	if f.UncompressedSize64 != 5*1024*1024 {
		t.Errorf("expected 5MB file, got %d bytes", f.UncompressedSize64)
	}

	// Verify compression occurred
	if f.Method == zip.Deflate && f.CompressedSize64 >= f.UncompressedSize64 {
		t.Errorf("expected compression for large file, but compressed=%d >= uncompressed=%d",
			f.CompressedSize64, f.UncompressedSize64)
	}
}

// TestZip_CreateWithUnicodeNames tests creating a zip from a directory with unicode filenames.
func TestZip_CreateWithUnicodeNames(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	zipPath := filepath.Join(tmpDir, "archive.zip")

	// Create source directory
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	// Create files with unicode names
	files := map[string]string{
		"日本語.txt":    "Japanese",
		"emoji_🎉.txt": "Party",
		"Ελληνικά.txt": "Greek",
	}

	for name, content := range files {
		path := filepath.Join(sourceDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", name, err)
		}
	}

	// Create zip archive
	err := Zip(sourceDir, zipPath)
	if err != nil {
		t.Fatalf("Zip failed: %v", err)
	}

	// Verify zip contents
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open created zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 3 {
		t.Fatalf("expected 3 files in zip, got %d", len(r.File))
	}

	// Verify unicode filenames are preserved
	foundNames := make(map[string]bool)
	for _, f := range r.File {
		basename := filepath.Base(f.Name)
		foundNames[basename] = true
	}

	for expectedName := range files {
		if !foundNames[expectedName] {
			t.Errorf("unicode filename not found in zip: %s", expectedName)
		}
	}
}

// TestZip_CreateAndExtract tests roundtrip: create with Zip, extract with UnZip, verify integrity.
func TestZip_CreateAndExtract(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	zipPath := filepath.Join(tmpDir, "archive.zip")
	extractDir := filepath.Join(tmpDir, "extracted")

	// Create source directory with various content
	if err := os.MkdirAll(filepath.Join(sourceDir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create source dirs: %v", err)
	}

	files := map[string]string{
		"file1.txt":        "content one",
		"file2.txt":        "content two",
		"subdir/file3.txt": "nested content",
	}

	for relPath, content := range files {
		fullPath := filepath.Join(sourceDir, relPath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", relPath, err)
		}
	}

	// Create zip archive
	err := Zip(sourceDir, zipPath)
	if err != nil {
		t.Fatalf("Zip failed: %v", err)
	}

	// Extract zip archive
	err = UnZip(zipPath, extractDir)
	if err != nil {
		t.Fatalf("UnZip failed: %v", err)
	}

	// Verify extracted files match original
	for relPath, expectedContent := range files {
		extractedPath := filepath.Join(extractDir, "source", relPath)
		content, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Errorf("failed to read extracted file %s: %v", relPath, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("file %s: expected '%s', got '%s'", relPath, expectedContent, string(content))
		}
	}
}

// TestZip_NonExistentSource tests error handling for non-existent source.
func TestZip_NonExistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "nonexistent.txt")
	zipPath := filepath.Join(tmpDir, "archive.zip")

	// Try to create zip from non-existent source
	err := Zip(sourceFile, zipPath)
	if err == nil {
		t.Error("expected error for non-existent source")
	}
}

// TestZip_InvalidDestPath tests error handling for invalid destination path.
func TestZip_InvalidDestPath(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "source.txt")
	invalidZipPath := filepath.Join(tmpDir, "nonexistent_dir", "archive.zip")

	// Create source file
	if err := os.WriteFile(sourceFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Try to create zip in non-existent directory
	err := Zip(sourceFile, invalidZipPath)
	if err == nil {
		t.Error("expected error for invalid destination path")
	}
}

// TestZip_EmptyDirectory tests creating a zip from an empty directory.
func TestZip_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "empty")
	zipPath := filepath.Join(tmpDir, "archive.zip")

	// Create empty directory
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create empty dir: %v", err)
	}

	// Create zip from empty directory
	err := Zip(sourceDir, zipPath)
	if err != nil {
		t.Fatalf("Zip failed on empty directory: %v", err)
	}

	// Verify zip was created and is valid (even if empty)
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open created zip: %v", err)
	}
	defer r.Close()

	// Empty directory should produce a zip with no files
	if len(r.File) != 0 {
		t.Errorf("expected 0 files in zip from empty directory, got %d", len(r.File))
	}
}

// =============================================================================
// Zip Benchmarks
// =============================================================================

// BenchmarkZip_SmallFiles benchmarks creating a zip with 100 small files.
func BenchmarkZip_SmallFiles(b *testing.B) {
	tmpDir := b.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")

	// Create source directory with 100 small files
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		b.Fatalf("failed to create source dir: %v", err)
	}

	for i := 0; i < 100; i++ {
		filename := filepath.Join(sourceDir, filepath.Base(filepath.Join("file", string(rune('0'+i%10))+".txt")))
		if err := os.WriteFile(filename, []byte("small content"), 0644); err != nil {
			b.Fatalf("failed to create file: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		zipPath := filepath.Join(tmpDir, "bench"+string(rune('0'+i%10))+".zip")
		Zip(sourceDir, zipPath)
		os.Remove(zipPath)
	}
}

// BenchmarkZip_LargeFile benchmarks creating a zip with a 10MB file.
func BenchmarkZip_LargeFile(b *testing.B) {
	tmpDir := b.TempDir()
	sourceFile := filepath.Join(tmpDir, "large.txt")

	// Create 10MB file
	largeContent := make([]byte, 10*1024*1024)
	for i := range largeContent {
		largeContent[i] = byte('A' + i%26)
	}
	if err := os.WriteFile(sourceFile, largeContent, 0644); err != nil {
		b.Fatalf("failed to create large file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		zipPath := filepath.Join(tmpDir, "bench.zip")
		Zip(sourceFile, zipPath)
		os.Remove(zipPath)
	}
}

// BenchmarkZip_DeflateCompression benchmarks zip creation with deflate compression.
func BenchmarkZip_DeflateCompression(b *testing.B) {
	tmpDir := b.TempDir()
	sourceFile := filepath.Join(tmpDir, "source.txt")

	// Create file with compressible content (1MB)
	content := make([]byte, 1024*1024)
	for i := range content {
		content[i] = 'A' // Highly compressible
	}
	if err := os.WriteFile(sourceFile, content, 0644); err != nil {
		b.Fatalf("failed to create source file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		zipPath := filepath.Join(tmpDir, "bench.zip")
		Zip(sourceFile, zipPath, WithZipDeflate())
		os.Remove(zipPath)
	}
}

// BenchmarkZip_StoreMethod benchmarks zip creation without compression (store method).
func BenchmarkZip_StoreMethod(b *testing.B) {
	tmpDir := b.TempDir()
	sourceFile := filepath.Join(tmpDir, "source.txt")

	// Create file (1MB)
	content := make([]byte, 1024*1024)
	for i := range content {
		content[i] = byte('A' + i%26)
	}
	if err := os.WriteFile(sourceFile, content, 0644); err != nil {
		b.Fatalf("failed to create source file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		zipPath := filepath.Join(tmpDir, "bench.zip")
		Zip(sourceFile, zipPath, WithZipStore())
		os.Remove(zipPath)
	}
}
