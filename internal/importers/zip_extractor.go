package importers

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"portal64api/internal/config"
	"strings"
	"time"

	"github.com/yeka/zip"
)

// ZIPExtractor handles extraction of password-protected ZIP files
type ZIPExtractor struct {
	config *config.ZIPConfig
	logger *log.Logger
}

// NewZIPExtractor creates a new ZIP extractor instance
func NewZIPExtractor(config *config.ZIPConfig, logger *log.Logger) *ZIPExtractor {
	return &ZIPExtractor{
		config: config,
		logger: logger,
	}
}

// ExtractFile extracts a single ZIP file to the specified directory
func (ze *ZIPExtractor) ExtractFile(zipPath, extractDir string) ([]string, error) {
	ze.logger.Printf("Extracting ZIP file: %s -> %s", zipPath, extractDir)

	// Ensure extract directory exists
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create extract directory: %w", err)
	}

	// Open ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer reader.Close()

	var extractedFiles []string
	totalFiles := len(reader.File)
	ze.logger.Printf("ZIP contains %d files", totalFiles)

	// Extract each file
	for i, file := range reader.File {
		extractedFile, err := ze.extractSingleFile(file, extractDir)
		if err != nil {
			return extractedFiles, fmt.Errorf("failed to extract %s: %w", file.Name, err)
		}

		if extractedFile != "" {
			extractedFiles = append(extractedFiles, extractedFile)
		}

		// Report progress
		progress := ((i + 1) * 100) / totalFiles
		if progress%20 == 0 || i == totalFiles-1 {
			ze.logger.Printf("Extraction progress: %d%% (%d/%d files)", progress, i+1, totalFiles)
		}
	}

	ze.logger.Printf("Successfully extracted %d files from %s", len(extractedFiles), filepath.Base(zipPath))
	return extractedFiles, nil
}

// ExtractFiles extracts multiple ZIP files
func (ze *ZIPExtractor) ExtractFiles(zipPaths []string, extractDir string) (map[string][]string, error) {
	if len(zipPaths) == 0 {
		return nil, fmt.Errorf("no ZIP files to extract")
	}

	results := make(map[string][]string)

	for _, zipPath := range zipPaths {
		extractedFiles, err := ze.ExtractFile(zipPath, extractDir)
		if err != nil {
			return results, fmt.Errorf("failed to extract %s: %w", zipPath, err)
		}
		results[zipPath] = extractedFiles
	}

	return results, nil
}

// extractSingleFile extracts a single file from the ZIP archive
func (ze *ZIPExtractor) extractSingleFile(file *zip.File, extractDir string) (string, error) {
	// Skip directories
	if file.FileInfo().IsDir() {
		return "", nil
	}

	// Sanitize file path to prevent zip slip attacks
	cleanPath := filepath.Join(extractDir, file.Name)
	if !strings.HasPrefix(cleanPath, filepath.Clean(extractDir)+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid file path: %s", file.Name)
	}

	// Set password if file is encrypted
	if file.IsEncrypted() {
		file.SetPassword(ze.config.Password)
	}

	// Open file in ZIP
	reader, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file in ZIP: %w", err)
	}
	defer reader.Close()

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create output file
	outFile, err := os.Create(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Copy data with progress tracking
	written, err := ze.copyWithProgress(outFile, reader, file.UncompressedSize64, file.Name)
	if err != nil {
		return "", fmt.Errorf("failed to copy file data: %w", err)
	}

	ze.logger.Printf("Extracted: %s (%d bytes)", file.Name, written)
	return cleanPath, nil
}

// copyWithProgress copies data with progress reporting for large files
func (ze *ZIPExtractor) copyWithProgress(dst io.Writer, src io.Reader, totalSize uint64, filename string) (int64, error) {
	const bufferSize = 32 * 1024 // 32KB buffer
	buffer := make([]byte, bufferSize)
	
	var written int64
	lastProgress := 0
	
	for {
		nr, er := src.Read(buffer)
		if nr > 0 {
			nw, ew := dst.Write(buffer[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = fmt.Errorf("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, fmt.Errorf("short write")
			}
			
			// Report progress for large files (>10MB)
			if totalSize > 10*1024*1024 {
				progress := int((written * 100) / int64(totalSize))
				if progress >= lastProgress+25 {
					ze.logger.Printf("Extract progress for %s: %d%% (%d/%d bytes)", 
						filename, progress, written, totalSize)
					lastProgress = progress
				}
			}
		}
		if er != nil {
			if er != io.EOF {
				return written, er
			}
			break
		}
	}
	return written, nil
}

// ValidateZIPFile validates that a ZIP file can be opened and contains expected content
func (ze *ZIPExtractor) ValidateZIPFile(zipPath string) error {
	ze.logger.Printf("Validating ZIP file: %s", zipPath)

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open ZIP file for validation: %w", err)
	}
	defer reader.Close()

	if len(reader.File) == 0 {
		return fmt.Errorf("ZIP file is empty")
	}

	// Check if any files are encrypted and we have password
	hasEncryptedFiles := false
	for _, file := range reader.File {
		if file.IsEncrypted() {
			hasEncryptedFiles = true
			break
		}
	}

	if hasEncryptedFiles && ze.config.Password == "" {
		return fmt.Errorf("ZIP file contains encrypted files but no password provided")
	}

	// Try to read first file to validate password
	if hasEncryptedFiles {
		for _, file := range reader.File {
			if file.IsEncrypted() && !file.FileInfo().IsDir() {
				file.SetPassword(ze.config.Password)
				testReader, err := file.Open()
				if err != nil {
					return fmt.Errorf("failed to decrypt file with provided password: %w", err)
				}
				
				// Try to read a few bytes
				buffer := make([]byte, 100)
				_, err = testReader.Read(buffer)
				testReader.Close()
				
				if err != nil && err != io.EOF {
					return fmt.Errorf("password validation failed: %w", err)
				}
				break // Only test one encrypted file
			}
		}
	}

	ze.logger.Printf("ZIP file validation successful: %d files, encrypted: %t", 
		len(reader.File), hasEncryptedFiles)
	return nil
}

// GetZIPInfo returns information about the ZIP file contents
func (ze *ZIPExtractor) GetZIPInfo(zipPath string) (*ZIPInfo, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer reader.Close()

	info := &ZIPInfo{
		TotalFiles:       len(reader.File),
		UncompressedSize: 0,
		CompressedSize:   0,
		HasEncryptedFiles: false,
		Files:           make([]ZIPFileInfo, 0, len(reader.File)),
	}

	for _, file := range reader.File {
		info.UncompressedSize += int64(file.UncompressedSize64)
		info.CompressedSize += int64(file.CompressedSize64)
		
		if file.IsEncrypted() {
			info.HasEncryptedFiles = true
		}

		if !file.FileInfo().IsDir() {
			info.Files = append(info.Files, ZIPFileInfo{
				Name:             file.Name,
				Size:             int64(file.UncompressedSize64),
				CompressedSize:   int64(file.CompressedSize64),
				ModTime:          file.ModTime(),
				IsEncrypted:      file.IsEncrypted(),
				CompressionMethod: file.Method,
			})
		}
	}

	return info, nil
}

// FindDatabaseDumps finds SQL dump files in the extracted directory
func (ze *ZIPExtractor) FindDatabaseDumps(extractDir string) (map[string]string, error) {
	dumps := make(map[string]string)

	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		filename := info.Name()
		
		// Look for SQL dump files
		if strings.HasSuffix(strings.ToLower(filename), ".sql") {
			database := ze.identifyDatabase(filename)
			if database != "" {
				dumps[database] = path
				ze.logger.Printf("Found database dump: %s -> %s", database, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search for database dumps: %w", err)
	}

	return dumps, nil
}

// identifyDatabase identifies which database a dump file belongs to
func (ze *ZIPExtractor) identifyDatabase(filename string) string {
	filename = strings.ToLower(filename)
	
	if strings.Contains(filename, "mvdsb") {
		return "mvdsb"
	}
	if strings.Contains(filename, "portal64_bdw") || strings.Contains(filename, "portal64bdw") {
		return "portal64_bdw"
	}
	
	return ""
}

// CleanupExtracted removes extracted files and directories
func (ze *ZIPExtractor) CleanupExtracted(extractDir string) error {
	ze.logger.Printf("Cleaning up extracted files in: %s", extractDir)
	
	if err := os.RemoveAll(extractDir); err != nil {
		return fmt.Errorf("failed to cleanup extracted files: %w", err)
	}
	
	return nil
}

// ZIPInfo contains information about a ZIP file
type ZIPInfo struct {
	TotalFiles        int           `json:"total_files"`
	UncompressedSize  int64         `json:"uncompressed_size"`
	CompressedSize    int64         `json:"compressed_size"`
	HasEncryptedFiles bool          `json:"has_encrypted_files"`
	Files             []ZIPFileInfo `json:"files"`
}

// ZIPFileInfo contains information about a file within a ZIP
type ZIPFileInfo struct {
	Name              string    `json:"name"`
	Size              int64     `json:"size"`
	CompressedSize    int64     `json:"compressed_size"`
	ModTime           time.Time `json:"mod_time"`
	IsEncrypted       bool      `json:"is_encrypted"`
	CompressionMethod uint16    `json:"compression_method"`
}

// TestPassword tests if the provided password can decrypt encrypted files
func (ze *ZIPExtractor) TestPassword(zipPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer reader.Close()

	// Find first encrypted file
	for _, file := range reader.File {
		if file.IsEncrypted() && !file.FileInfo().IsDir() {
			file.SetPassword(ze.config.Password)
			testReader, err := file.Open()
			if err != nil {
				return fmt.Errorf("password test failed: %w", err)
			}
			testReader.Close()
			return nil // Password works
		}
	}

	return fmt.Errorf("no encrypted files found to test password")
}
