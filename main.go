// MDOnline - Zero-build Markdown documentation site launcher
// All static files are embedded into the executable.
// Double-click to run: auto-init docs, generate sidebar, start server, open browser.
package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

//go:embed all:static
var staticFiles embed.FS

// ── Configuration ──

var groupOrder = []string{"快速开始", "功能特性", "写作指南", "部署运维"}

var groupNames = map[string]string{
	"快速开始": "快速开始",
	"功能特性": "功能特性",
	"写作指南": "写作指南",
	"部署运维": "部署运维",
}

var skipFiles = map[string]bool{
	"_sidebar.md": true,
	"_navbar.md":  true,
	"README.md":   true,
}

const listenPort = ":8080"

// ── Initialization: extract embedded files if missing ──

func initWorkDir(baseDir string) {
	docsDir := filepath.Join(baseDir, "docs")
	imagesDir := filepath.Join(baseDir, "images")

	// Check if docs/ already exists
	needsInit := false
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		needsInit = true
	}

	if needsInit {
		fmt.Println("First run: initializing docs directory...")
	} else {
		// Even if docs/ exists, check if core files are missing
		coreFiles := []string{"index.html", "style.css", "vue.css", "docsify.min.js", "search.min.js", "favicon.svg"}
		for _, f := range coreFiles {
			if _, err := os.Stat(filepath.Join(baseDir, f)); os.IsNotExist(err) {
				needsInit = true
				break
			}
		}
	}

	if !needsInit {
		return
	}

	// Extract all embedded files to baseDir
	fs.WalkDir(staticFiles, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		// path = "static/index.html", "static/docs/README.md", etc.
		relPath := strings.TrimPrefix(path, "static/")
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			return nil
		}

		localPath := filepath.Join(baseDir, relPath)

		// Skip if file already exists (don't overwrite user edits)
		if _, err := os.Stat(localPath); err == nil {
			return nil
		}

		// Create parent directory
		os.MkdirAll(filepath.Dir(localPath), 0755)

		// Read embedded file and write to disk
		data, err := staticFiles.ReadFile(path)
		if err != nil {
			fmt.Printf("  Warning: cannot read embedded %s: %v\n", relPath, err)
			return nil
		}
		if err := os.WriteFile(localPath, data, 0644); err != nil {
			fmt.Printf("  Warning: cannot write %s: %v\n", localPath, err)
		} else {
			fmt.Printf("  Created: %s\n", relPath)
		}
		return nil
	})

	_ = imagesDir // images dir is created by the walk above if needed
	fmt.Println("Initialization complete.")
}

// ── Sidebar generation ──

var headingRe = regexp.MustCompile(`^#\s+(.+)`)

func getFirstHeading(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if m := headingRe.FindStringSubmatch(line); m != nil {
			return strings.TrimSpace(m[1])
		}
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
			break
		}
	}
	return ""
}

func generateSidebar(baseDir string) string {
	docsDir := filepath.Join(baseDir, "docs")

	entries, err := os.ReadDir(docsDir)
	if err != nil {
		fmt.Printf("Warning: cannot read docs/ directory: %v\n", err)
		return ""
	}

	// Collect subdirectories
	var subdirs []string
	for _, e := range entries {
		if e.IsDir() {
			subdirs = append(subdirs, e.Name())
		}
	}

	// Sort by groupOrder, then alphabetically
	orderSet := map[string]bool{}
	for _, g := range groupOrder {
		orderSet[g] = true
	}
	var ordered []string
	for _, g := range groupOrder {
		for _, d := range subdirs {
			if d == g {
				ordered = append(ordered, d)
			}
		}
	}
	var remaining []string
	for _, d := range subdirs {
		if !orderSet[d] {
			remaining = append(remaining, d)
		}
	}
	sort.Strings(remaining)
	allDirs := append(ordered, remaining...)

	var lines []string
	for _, dirname := range allDirs {
		dirPath := filepath.Join(docsDir, dirname)
		displayName := dirname
		if n, ok := groupNames[dirname]; ok {
			displayName = n
		}

		fileEntries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}

		type mdEntry struct {
			text string
			url  string
		}
		var mdFiles []mdEntry

		for _, fe := range fileEntries {
			name := fe.Name()
			if !strings.HasSuffix(name, ".md") || skipFiles[name] {
				continue
			}
			linkName := strings.TrimSuffix(name, ".md")
			heading := getFirstHeading(filepath.Join(dirPath, name))
			linkText := linkName
			if heading != "" {
				linkText = heading
			}
			linkURL := "/docs/" + dirname + "/" + linkName
			mdFiles = append(mdFiles, mdEntry{linkText, linkURL})
		}

		if len(mdFiles) == 0 {
			continue
		}

		lines = append(lines, "- "+displayName)
		for _, entry := range mdFiles {
			lines = append(lines, fmt.Sprintf("  - [%s](%s)", entry.text, entry.url))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func writeSidebar(baseDir, content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("no .md files found in docs/")
	}

	// Write root _sidebar.md
	rootSidebar := filepath.Join(baseDir, "_sidebar.md")
	if err := os.WriteFile(rootSidebar, []byte(content+"\n"), 0644); err != nil {
		return fmt.Errorf("write root _sidebar.md: %w", err)
	}

	// Write docs/_sidebar.md
	docsSidebar := filepath.Join(baseDir, "docs", "_sidebar.md")
	if err := os.WriteFile(docsSidebar, []byte(content+"\n"), 0644); err != nil {
		return fmt.Errorf("write docs/_sidebar.md: %w", err)
	}

	// Remove sub-directory _sidebar.md files
	docsDir := filepath.Join(baseDir, "docs")
	entries, _ := os.ReadDir(docsDir)
	for _, e := range entries {
		if e.IsDir() {
			subSidebar := filepath.Join(docsDir, e.Name(), "_sidebar.md")
			os.Remove(subSidebar)
		}
	}

	return nil
}

// ── HTTP server ──

func main() {
	// Determine base directory (where the exe lives)
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error: cannot determine executable path")
		os.Exit(1)
	}
	baseDir := filepath.Dir(exePath)

	// Initialize: extract embedded files if local files are missing
	initWorkDir(baseDir)

	// Generate sidebar on startup
	fmt.Println("Generating sidebar...")
	content := generateSidebar(baseDir)
	if err := writeSidebar(baseDir, content); err != nil {
		fmt.Printf("Warning: sidebar generation failed: %v\n", err)
	} else {
		fmt.Println("Sidebar generated.")
	}

	mux := http.NewServeMux()

	// Refresh endpoint: regenerate sidebar and return OK
	mux.HandleFunc("/__refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		content := generateSidebar(baseDir)
		if err := writeSidebar(baseDir, content); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, "ok")
		fmt.Println("Sidebar refreshed.")
	})

	// Static file serving: read file directly (full control over Chinese paths)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// URL-decode the path for Chinese filenames
		reqPath := r.URL.Path
		if decoded, err := url.PathUnescape(reqPath); err == nil {
			reqPath = decoded
		}

		// Normalize: remove leading slash, convert to platform path
		relPath := strings.TrimPrefix(reqPath, "/")
		if relPath == "" {
			relPath = "index.html"
		}

		// Try reading local file first
		localPath := filepath.Join(baseDir, filepath.FromSlash(relPath))
		if data, err := os.ReadFile(localPath); err == nil {
			fmt.Printf("[200] local  %s\n", relPath)
			w.Header().Set("Content-Type", mimeType(relPath))
			w.Write(data)
			return
		}

		// If no extension, try appending .md (Docsify requests files without .md)
		if filepath.Ext(relPath) == "" {
			mdRelPath := relPath + ".md"
			mdLocalPath := filepath.Join(baseDir, filepath.FromSlash(mdRelPath))
			if data, err := os.ReadFile(mdLocalPath); err == nil {
				fmt.Printf("[200] local  %s -> %s\n", relPath, mdRelPath)
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Write(data)
				return
			}
			// Also try embedded .md
			mdEmbedPath := "static/" + filepath.ToSlash(mdRelPath)
			if data, err := staticFiles.ReadFile(mdEmbedPath); err == nil {
				fmt.Printf("[200] embed  %s -> %s\n", relPath, mdRelPath)
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Write(data)
				return
			}
		}

		// Fallback to embedded file
		embedPath := "static/" + filepath.ToSlash(relPath)
		if data, err := staticFiles.ReadFile(embedPath); err == nil {
			fmt.Printf("[200] embed  %s\n", relPath)
			w.Header().Set("Content-Type", mimeType(relPath))
			w.Write(data)
			return
		}

		// SPA fallback: ONLY for routes without file extension (Docsify hash routing)
		// Files with extensions (.md, .png, etc.) that are not found should return 404
		// so Docsify can walk up the directory tree to find _sidebar.md
		if filepath.Ext(relPath) != "" {
			fmt.Printf("[404] %s\n", relPath)
			http.NotFound(w, r)
			return
		}

		fmt.Printf("[SPA] fallback %s -> index.html\n", relPath)
		if data, err := os.ReadFile(filepath.Join(baseDir, "index.html")); err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(data)
			return
		}
		if data, err := staticFiles.ReadFile("static/index.html"); err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(data)
			return
		}

		fmt.Printf("[404] %s\n", relPath)
		http.NotFound(w, r)
	})

	// Open browser
	url := "http://localhost" + listenPort
	go openBrowser(url)

	fmt.Printf("MDOnline running at %s\n", url)
	fmt.Println("Press Ctrl+C to stop.")
	if err := http.ListenAndServe(listenPort, mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func mimeType(path string) string {
	switch {
	case strings.HasSuffix(path, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(path, ".md"):
		return "text/plain; charset=utf-8"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(path, ".json"):
		return "application/json; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}
