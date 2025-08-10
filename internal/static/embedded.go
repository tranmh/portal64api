package static

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Embed the demo directory content
//go:embed demo
var DemoFiles embed.FS

// Embed swagger UI assets
//go:embed swagger
var SwaggerFiles embed.FS

// DebugEmbeddedFiles prints all embedded files for debugging (commented out for production)
/*
func DebugEmbeddedFiles() {
	fmt.Println("=== DEBUG: Embedded Demo Files ===")
	err := fs.WalkDir(DemoFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			fmt.Printf("[DIR]  %s\n", path)
		} else {
			fmt.Printf("[FILE] %s\n", path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking demo files: %v\n", err)
	}
	
	fmt.Println("=== DEBUG: Embedded Swagger Files ===")
	err = fs.WalkDir(SwaggerFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			fmt.Printf("[DIR]  %s\n", path)
		} else {
			fmt.Printf("[FILE] %s\n", path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking swagger files: %v\n", err)
	}
}
*/

// EmbeddedFileSystem implements gin's FileSystem interface for embedded files
type EmbeddedFileSystem struct {
	http.FileSystem
}

func (e EmbeddedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	return err == nil
}

// NewEmbeddedFileSystem creates a new embedded filesystem for serving static files
func NewEmbeddedFileSystem(fsys embed.FS, root string) *EmbeddedFileSystem {
	sub, err := fs.Sub(fsys, root)
	if err != nil {
		panic(err)
	}
	return &EmbeddedFileSystem{
		FileSystem: http.FS(sub),
	}
}

// ServeDemoFile serves individual demo files from embedded filesystem
func ServeDemoFile(c *gin.Context) {
	requestedPath := c.Param("filepath")
	
	// Handle root path requests (/demo/ or /demo)
	if requestedPath == "" || requestedPath == "/" {
		requestedPath = "index.html"
	} else {
		// Remove leading slash if present
		requestedPath = strings.TrimPrefix(requestedPath, "/")
	}
	
	// Clean the path to prevent directory traversal - use path.Clean for embedded FS
	cleanPath := path.Clean(requestedPath)
	if strings.HasPrefix(cleanPath, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path"})
		return
	}
	
	// Try to read the file from embedded filesystem
	fullPath := "demo/" + cleanPath
	content, err := DemoFiles.ReadFile(fullPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	
	// Set appropriate content type based on file extension
	ext := filepath.Ext(cleanPath)
	switch ext {
	case ".html":
		c.Header("Content-Type", "text/html; charset=utf-8")
	case ".css":
		c.Header("Content-Type", "text/css")
	case ".js":
		c.Header("Content-Type", "application/javascript")
	case ".json":
		c.Header("Content-Type", "application/json")
	default:
		c.Header("Content-Type", "text/plain")
	}
	
	c.Data(http.StatusOK, c.GetHeader("Content-Type"), content)
}

// ServeSwaggerUI serves the embedded Swagger UI
func ServeSwaggerUI(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Portal64 API Documentation</title>
    <link rel="stylesheet" type="text/css" href="/swagger/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="/swagger/swagger-ui-bundle.js"></script>
    <script src="/swagger/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger/doc.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ServeSwaggerAsset serves embedded swagger UI assets (CSS, JS)
func ServeSwaggerAsset(c *gin.Context) {
	filename := c.Param("filename")
	
	// Clean the filename to prevent directory traversal - use path.Clean for embedded FS
	filename = path.Clean(filename)
	if strings.HasPrefix(filename, "..") || strings.Contains(filename, "/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filename"})
		return
	}
	
	// Try to read the file from embedded filesystem
	content, err := SwaggerFiles.ReadFile("swagger/" + filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}
	
	// Set appropriate content type
	ext := filepath.Ext(filename)
	switch ext {
	case ".css":
		c.Header("Content-Type", "text/css")
	case ".js":
		c.Header("Content-Type", "application/javascript")
	default:
		c.Header("Content-Type", "text/plain")
	}
	
	c.Data(http.StatusOK, c.GetHeader("Content-Type"), content)
}
