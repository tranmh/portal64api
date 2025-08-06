package importers

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"portal64api/internal/config"
	"portal64api/internal/models"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SCPDownloader handles downloading files via SCP/SFTP
type SCPDownloader struct {
	config *config.SCPConfig
	logger *log.Logger
}

// NewSCPDownloader creates a new SCP downloader instance
func NewSCPDownloader(config *config.SCPConfig, logger *log.Logger) *SCPDownloader {
	return &SCPDownloader{
		config: config,
		logger: logger,
	}
}

// ListFiles lists remote files matching the configured patterns
func (sd *SCPDownloader) ListFiles() ([]models.FileMetadata, error) {
	sd.logger.Printf("Connecting to %s:%d", sd.config.Host, sd.config.Port)
	
	client, sftpClient, err := sd.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()
	defer sftpClient.Close()

	var allFiles []models.FileMetadata

	// List files in remote directory
	files, err := sftpClient.ReadDir(sd.config.RemotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote directory %s: %w", sd.config.RemotePath, err)
	}

	// Filter files by patterns
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		for _, pattern := range sd.config.FilePatterns {
			if sd.matchesPattern(filename, pattern) {
				// Extract database name from pattern/filename
				database := sd.extractDatabaseName(filename, pattern)
				
				fileMetadata := models.FileMetadata{
					Filename: filename,
					Size:     file.Size(),
					ModTime:  file.ModTime(),
					Pattern:  pattern,
					Database: database,
				}

				allFiles = append(allFiles, fileMetadata)
				sd.logger.Printf("Found file: %s (size: %d, modified: %s)", 
					filename, file.Size(), file.ModTime().Format(time.RFC3339))
				break // Don't match the same file to multiple patterns
			}
		}
	}

	if len(allFiles) == 0 {
		return nil, fmt.Errorf("no files found matching patterns: %v", sd.config.FilePatterns)
	}

	sd.logger.Printf("Found %d files matching patterns", len(allFiles))
	return allFiles, nil
}

// DownloadFiles downloads the specified files to the local directory
func (sd *SCPDownloader) DownloadFiles(files []models.FileMetadata, localDir string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to download")
	}

	// Ensure local directory exists
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	sd.logger.Printf("Connecting to %s:%d for download", sd.config.Host, sd.config.Port)

	client, sftpClient, err := sd.connect()
	if err != nil {
		return fmt.Errorf("failed to connect for download: %w", err)
	}
	defer client.Close()
	defer sftpClient.Close()

	for _, file := range files {
		if err := sd.downloadFile(sftpClient, file, localDir); err != nil {
			return fmt.Errorf("failed to download %s: %w", file.Filename, err)
		}
	}

	sd.logger.Printf("Successfully downloaded %d files", len(files))
	return nil
}

// downloadFile downloads a single file
func (sd *SCPDownloader) downloadFile(sftpClient *sftp.Client, file models.FileMetadata, localDir string) error {
	remotePath := path.Join(sd.config.RemotePath, file.Filename)
	localPath := filepath.Join(localDir, file.Filename)

	sd.logger.Printf("Downloading %s -> %s", remotePath, localPath)

	// Open remote file
	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer remoteFile.Close()

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	// Copy with progress tracking
	written, err := sd.copyWithProgress(localFile, remoteFile, file.Size, file.Filename)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	if written != file.Size {
		return fmt.Errorf("incomplete download: expected %d bytes, got %d", file.Size, written)
	}

	// Verify file integrity by size
	stat, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file: %w", err)
	}

	if stat.Size() != file.Size {
		return fmt.Errorf("file size mismatch: expected %d, got %d", file.Size, stat.Size())
	}

	sd.logger.Printf("Successfully downloaded %s (%d bytes)", file.Filename, written)
	return nil
}

// copyWithProgress copies data with progress reporting
func (sd *SCPDownloader) copyWithProgress(dst io.Writer, src io.Reader, totalSize int64, filename string) (int64, error) {
	const bufferSize = 64 * 1024 // 64KB buffer
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
			
			// Report progress every 10%
			if totalSize > 0 {
				progress := int((written * 100) / totalSize)
				if progress >= lastProgress+10 {
					sd.logger.Printf("Download progress for %s: %d%% (%d/%d bytes)", 
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

// connect establishes SSH and SFTP connections
func (sd *SCPDownloader) connect() (*ssh.Client, *sftp.Client, error) {
	// SSH client configuration
	config := &ssh.ClientConfig{
		User: sd.config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(sd.config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: In production, use proper host key verification
		Timeout:         sd.config.Timeout,
	}

	// Connect to SSH server
	addr := net.JoinHostPort(sd.config.Host, strconv.Itoa(sd.config.Port))
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	// Create SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return sshClient, sftpClient, nil
}

// matchesPattern checks if filename matches the given pattern
func (sd *SCPDownloader) matchesPattern(filename, pattern string) bool {
	if pattern == "" {
		return false
	}

	// Simple wildcard matching
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) != 2 {
			return false // Only support single * for now
		}

		prefix := parts[0]
		suffix := parts[1]

		return strings.HasPrefix(filename, prefix) && strings.HasSuffix(filename, suffix)
	}

	return filename == pattern
}

// extractDatabaseName extracts database name from filename and pattern
func (sd *SCPDownloader) extractDatabaseName(filename, pattern string) string {
	// Try to extract database name from pattern
	if strings.Contains(pattern, "mvdsb") {
		return "mvdsb"
	}
	if strings.Contains(pattern, "portal64_bdw") {
		return "portal64_bdw"
	}

	// Fallback: try to extract from filename
	if strings.Contains(filename, "mvdsb") {
		return "mvdsb"
	}
	if strings.Contains(filename, "portal64_bdw") {
		return "portal64_bdw"
	}

	return ""
}

// TestConnection tests the SCP connection
func (sd *SCPDownloader) TestConnection() error {
	sd.logger.Printf("Testing connection to %s:%d", sd.config.Host, sd.config.Port)

	client, sftpClient, err := sd.connect()
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer client.Close()
	defer sftpClient.Close()

	// Test directory access
	_, err = sftpClient.ReadDir(sd.config.RemotePath)
	if err != nil {
		return fmt.Errorf("failed to access remote directory %s: %w", sd.config.RemotePath, err)
	}

	sd.logger.Printf("Connection test successful")
	return nil
}

// CalculateChecksum calculates SHA256 checksum of a local file
func (sd *SCPDownloader) CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return fmt.Sprintf("sha256:%x", hasher.Sum(nil)), nil
}

// GetFileInfo gets information about a remote file
func (sd *SCPDownloader) GetFileInfo(filename string) (*models.FileMetadata, error) {
	client, sftpClient, err := sd.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()
	defer sftpClient.Close()

	remotePath := path.Join(sd.config.RemotePath, filename)
	stat, err := sftpClient.Stat(remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat remote file: %w", err)
	}

	return &models.FileMetadata{
		Filename: filename,
		Size:     stat.Size(),
		ModTime:  stat.ModTime(),
	}, nil
}
