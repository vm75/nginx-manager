package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

var configDir string

type FileInfo struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	IsDir      bool   `json:"isDir"`
	IsSymlink  bool   `json:"isSymlink"`
	LinkTarget string `json:"linkTarget,omitempty"`
	Size       int64  `json:"size"`
	ModTime    string `json:"modTime"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CertificateInfo struct {
	Domain      string `json:"domain"`
	Path        string `json:"path"`
	CertFile    string `json:"certFile"`
	KeyFile     string `json:"keyFile"`
	NotBefore   string `json:"notBefore"`
	NotAfter    string `json:"notAfter"`
	DaysLeft    int    `json:"daysLeft"`
	IsWildcard  bool   `json:"isWildcard"`
}

type ObtainCertRequest struct {
	Domains     []string          `json:"domains"`
	Email       string            `json:"email"`
	Challenge   string            `json:"challenge"` // "http-01" or "dns-01"
	Provider    string            `json:"provider"`  // "namesilo", "duckdns", "namecheap"
	Credentials map[string]string `json:"credentials"`
	Staging     bool              `json:"staging"` // Use Let's Encrypt staging server
	Force       bool              `json:"force"`   // Force renewal even if cert exists
}

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	flag.StringVar(&configDir, "config", "/etc/nginx", "Nginx config directory")
	flag.Parse()

	// Validate config directory
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		log.Fatalf("Config directory does not exist: %s", configDir)
	}

	configDir, _ = filepath.Abs(configDir)
	log.Printf("Starting nginx editor server on port %s", *port)
	log.Printf("Config directory: %s", configDir)

	// Setup routes
	http.HandleFunc("/api/files", handleFiles)
	http.HandleFunc("/api/file/read", handleFileRead)
	http.HandleFunc("/api/file/write", handleFileWrite)
	http.HandleFunc("/api/file/create", handleFileCreate)
	http.HandleFunc("/api/file/delete", handleFileDelete)
	http.HandleFunc("/api/file/rename", handleFileRename)
	http.HandleFunc("/api/file/move", handleFileMove)
	http.HandleFunc("/api/file/symlink", handleSymlinkCreate)
	http.HandleFunc("/api/nginx/test", handleNginxTest)
	http.HandleFunc("/api/nginx/reload", handleNginxReload)
	http.HandleFunc("/api/logs/access", handleAccessLog)
	http.HandleFunc("/api/logs/error", handleErrorLog)
	http.HandleFunc("/api/logs/cert-obtain", handleCertObtainLog)
	http.HandleFunc("/api/certificates", handleCertificates)
	http.HandleFunc("/api/certificates/obtain", handleObtainCertificate)
	http.HandleFunc("/api/certificates/delete", handleDeleteCertificate)

	// Serve frontend
	frontendDist, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.FS(frontendDist)))

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

// List files in directory
func handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	fullPath := filepath.Join(configDir, path)

	// Security check
	if !strings.HasPrefix(fullPath, configDir) {
		sendError(w, "Invalid path", http.StatusForbidden)
		return
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	files := make([]FileInfo, 0)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		relativePath := filepath.Join(path, entry.Name())
		entryFullPath := filepath.Join(fullPath, entry.Name())

		// Check if it's a symlink
		isSymlink := info.Mode()&os.ModeSymlink != 0
		linkTarget := ""
		if isSymlink {
			if target, err := os.Readlink(entryFullPath); err == nil {
				linkTarget = target
			}
		}

		files = append(files, FileInfo{
			Name:       entry.Name(),
			Path:       relativePath,
			IsDir:      entry.IsDir(),
			IsSymlink:  isSymlink,
			LinkTarget: linkTarget,
			Size:       info.Size(),
			ModTime:    info.ModTime().Format(time.RFC3339),
		})
	}

	sendJSON(w, files)
}

// Read file content
func handleFileRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		sendError(w, "Path is required", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(configDir, path)

	// Security check
	if !strings.HasPrefix(fullPath, configDir) {
		sendError(w, "Invalid path", http.StatusForbidden)
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

// Write file content
func handleFileWrite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(configDir, req.Path)

	// Security check
	if !strings.HasPrefix(fullPath, configDir) {
		sendError(w, "Invalid path", http.StatusForbidden)
		return
	}

	if err := os.WriteFile(fullPath, []byte(req.Content), 0644); err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{"status": "ok"})
}

// Create file or directory
func handleFileCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path  string `json:"path"`
		IsDir bool   `json:"isDir"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(configDir, req.Path)

	// Security check
	if !strings.HasPrefix(fullPath, configDir) {
		sendError(w, "Invalid path", http.StatusForbidden)
		return
	}

	if req.IsDir {
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			sendError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			sendError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := os.WriteFile(fullPath, []byte(""), 0644); err != nil {
			sendError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	sendJSON(w, map[string]string{"status": "ok"})
}

// Delete file or directory
func handleFileDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(configDir, req.Path)

	// Security check
	if !strings.HasPrefix(fullPath, configDir) {
		sendError(w, "Invalid path", http.StatusForbidden)
		return
	}

	if err := os.RemoveAll(fullPath); err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{"status": "ok"})
}

// Rename file or directory
func handleFileRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OldPath string `json:"oldPath"`
		NewPath string `json:"newPath"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	oldFullPath := filepath.Join(configDir, req.OldPath)
	newFullPath := filepath.Join(configDir, req.NewPath)

	// Security check
	if !strings.HasPrefix(oldFullPath, configDir) || !strings.HasPrefix(newFullPath, configDir) {
		sendError(w, "Invalid path", http.StatusForbidden)
		return
	}

	if err := os.Rename(oldFullPath, newFullPath); err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{"status": "ok"})
}

// Create symlink
func handleSymlinkCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		LinkPath   string `json:"linkPath"`
		TargetPath string `json:"targetPath"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	linkFullPath := filepath.Join(configDir, req.LinkPath)

	// Security check for link path
	if !strings.HasPrefix(linkFullPath, configDir) {
		sendError(w, "Invalid link path", http.StatusForbidden)
		return
	}

	// Target can be relative or absolute
	// If it starts with /, treat as absolute within config dir
	// Otherwise, treat as relative to the link's directory
	targetPath := req.TargetPath
	if strings.HasPrefix(targetPath, "/") {
		// Absolute path within config dir
		targetPath = filepath.Join(configDir, targetPath)
		// Convert back to relative path from link location
		linkDir := filepath.Dir(linkFullPath)
		relPath, err := filepath.Rel(linkDir, targetPath)
		if err != nil {
			sendError(w, err.Error(), http.StatusBadRequest)
			return
		}
		targetPath = relPath
	}

	// Create the symlink
	if err := os.Symlink(targetPath, linkFullPath); err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{"status": "ok"})
}

// Move file or directory
func handleFileMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SourcePath string `json:"sourcePath"`
		TargetPath string `json:"targetPath"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sourceFullPath := filepath.Join(configDir, req.SourcePath)
	targetFullPath := filepath.Join(configDir, req.TargetPath)

	// Security check
	if !strings.HasPrefix(sourceFullPath, configDir) || !strings.HasPrefix(targetFullPath, configDir) {
		sendError(w, "Invalid path", http.StatusForbidden)
		return
	}

	// If target is a directory, move source into it
	if info, err := os.Stat(targetFullPath); err == nil && info.IsDir() {
		targetFullPath = filepath.Join(targetFullPath, filepath.Base(sourceFullPath))
	}

	if err := os.Rename(sourceFullPath, targetFullPath); err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{"status": "ok"})
}

// Test nginx configuration
func handleNginxTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cmd := exec.Command("nginx", "-t")
	output, err := cmd.CombinedOutput()

	result := map[string]interface{}{
		"output": string(output),
		"success": err == nil,
	}

	sendJSON(w, result)
}

// Reload nginx
func handleNginxReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cmd := exec.Command("nginx", "-s", "reload")
	output, err := cmd.CombinedOutput()

	result := map[string]interface{}{
		"output": string(output),
		"success": err == nil,
	}

	sendJSON(w, result)
}

// Get access log
func handleAccessLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lines := r.URL.Query().Get("lines")
	if lines == "" {
		lines = "100"
	}

	logPath := findLogPath("access_log")
	content := readLastLines(logPath, lines)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(content))
}

// Get error log
func handleErrorLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lines := r.URL.Query().Get("lines")
	if lines == "" {
		lines = "100"
	}

	logPath := findLogPath("error_log")
	content := readLastLines(logPath, lines)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(content))
}

// Get certificate obtain log
func handleCertObtainLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lines := r.URL.Query().Get("lines")
	if lines == "" {
		lines = "500"
	}

	logPath := "/var/log/cert-obtain.log"
	content := readLastLines(logPath, lines)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(content))
}

// Delete certificate
func handleDeleteCertificate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CertFile string `json:"certFile"`
		KeyFile  string `json:"keyFile"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.CertFile == "" {
		sendError(w, "Certificate file path is required", http.StatusBadRequest)
		return
	}

	// Security check - ensure files are in the ssl directory
	sslDir := filepath.Join(configDir, "ssl")
	certPath := filepath.Clean(req.CertFile)
	if !strings.HasPrefix(certPath, sslDir) {
		sendError(w, "Certificate file must be in ssl directory", http.StatusForbidden)
		return
	}

	// Delete certificate file
	if err := os.Remove(req.CertFile); err != nil {
		if !os.IsNotExist(err) {
			sendError(w, "Failed to delete certificate: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Delete key file if provided
	if req.KeyFile != "" {
		keyPath := filepath.Clean(req.KeyFile)
		if !strings.HasPrefix(keyPath, sslDir) {
			sendError(w, "Key file must be in ssl directory", http.StatusForbidden)
			return
		}
		if err := os.Remove(req.KeyFile); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Failed to delete key file: %v", err)
		}
	}

	sendJSON(w, map[string]interface{}{
		"success": true,
		"message": "Certificate deleted successfully",
	})
}

// Find log path from nginx configuration
func findLogPath(logType string) string {
	// Default paths
	defaultPaths := map[string]string{
		"access_log": "/var/log/nginx/access.log",
		"error_log":  "/var/log/nginx/error.log",
	}

	// Try to find nginx.conf in config directory
	nginxConf := filepath.Join(configDir, "nginx.conf")
	if _, err := os.Stat(nginxConf); err != nil {
		// Try common nginx config locations
		nginxConf = "/etc/nginx/nginx.conf"
		if _, err := os.Stat(nginxConf); err != nil {
			return defaultPaths[logType]
		}
	}

	// Parse the config file
	logPath := parseLogFromConfig(nginxConf, logType)
	if logPath != "" {
		return logPath
	}

	return defaultPaths[logType]
}

// Parse log path from nginx config file
func parseLogFromConfig(configPath, logType string) string {
	file, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Regex to match log directives
	// access_log /path/to/file [format];
	// error_log /path/to/file [level];
	logRegex := regexp.MustCompile(`^\s*` + logType + `\s+([^\s;]+)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		matches := logRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			logPath := matches[1]

			// Skip special values
			if logPath == "off" || logPath == "syslog:" || strings.HasPrefix(logPath, "syslog:") {
				continue
			}

			// Handle relative paths
			if !filepath.IsAbs(logPath) {
				logPath = filepath.Join(filepath.Dir(configPath), logPath)
			}

			return logPath
		}
	}

	return ""
}

// List all certificates
func handleCertificates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	certsDir := filepath.Join(configDir, "ssl")
	certs := []CertificateInfo{}

	// Check if ssl directory exists
	if _, err := os.Stat(certsDir); os.IsNotExist(err) {
		sendJSON(w, certs)
		return
	}

	// Walk through ssl directory
	filepath.Walk(certsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Look for .crt or .pem files
		if strings.HasSuffix(path, ".crt") || strings.HasSuffix(path, ".pem") {
			certInfo := parseCertificate(path)
			if certInfo != nil {
				certs = append(certs, *certInfo)
			}
		}
		return nil
	})

	sendJSON(w, certs)
}

// Parse certificate file
func parseCertificate(certPath string) *CertificateInfo {
	// Use openssl to get certificate info
	cmd := exec.Command("openssl", "x509", "-in", certPath, "-noout", "-subject", "-dates")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	info := &CertificateInfo{
		Path:     certPath,
		CertFile: certPath,
	}

	// Find corresponding key file
	keyPath := strings.TrimSuffix(certPath, filepath.Ext(certPath)) + ".key"
	if _, err := os.Stat(keyPath); err == nil {
		info.KeyFile = keyPath
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "subject=") {
			// Extract domain from subject
			subjectParts := strings.Split(line, "CN = ")
			if len(subjectParts) > 1 {
				domain := strings.TrimSpace(subjectParts[1])
				info.Domain = domain
				info.IsWildcard = strings.HasPrefix(domain, "*.")
			}
		} else if strings.HasPrefix(line, "notBefore=") {
			info.NotBefore = strings.TrimPrefix(line, "notBefore=")
		} else if strings.HasPrefix(line, "notAfter=") {
			notAfter := strings.TrimPrefix(line, "notAfter=")
			info.NotAfter = notAfter

			// Calculate days left
			if t, err := time.Parse("Jan 2 15:04:05 2006 MST", notAfter); err == nil {
				daysLeft := int(time.Until(t).Hours() / 24)
				info.DaysLeft = daysLeft
			}
		}
	}

	return info
}

// Obtain certificate using acme.sh
func handleObtainCertificate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ObtainCertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Log file for certificate operations
	logFile := "/var/log/cert-obtain.log"
	logCertOp := func(message string) {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			f.WriteString(fmt.Sprintf("%s: %s\n", time.Now().Format(time.RFC3339), message))
		}
		log.Println(message) // Also log to stdout
	}

	logCertOp(fmt.Sprintf("========================================\nCertificate obtain request started - Domains: %v, Challenge: %s, Provider: %s, Staging: %v, Force: %v",
		req.Domains, req.Challenge, req.Provider, req.Staging, req.Force))

	// Validate request
	if len(req.Domains) == 0 || req.Email == "" {
		logCertOp("ERROR: At least one domain and email are required")
		sendError(w, "At least one domain and email are required", http.StatusBadRequest)
		return
	}

	if req.Challenge != "http-01" && req.Challenge != "dns-01" && req.Challenge != "tls-alpn-01" {
		sendError(w, "Challenge must be http-01, dns-01, or tls-alpn-01", http.StatusBadRequest)
		return
	}

	if req.Challenge == "dns-01" && req.Provider == "" {
		sendError(w, "DNS provider is required for dns-01 challenge", http.StatusBadRequest)
		return
	}

	// Prepare acme.sh directory
	acmeDir := "/root/.acme.sh"
	os.MkdirAll(acmeDir, 0755)

	// Prepare certificate output directory
	sslDir := filepath.Join(configDir, "ssl")
	os.MkdirAll(sslDir, 0755)

	// Build acme.sh command
	args := []string{
		"--issue",
		"--email", req.Email,
		"--debug", // Add debug output
	}

	// Add force flag if requested
	if req.Force {
		args = append(args, "--force")
		logCertOp("Force renewal enabled")
	}

	// Add all domains
	for _, domain := range req.Domains {
		args = append(args, "--domain", domain)
		logCertOp(fmt.Sprintf("Adding domain: %s", domain))
	}

	// Use staging server if requested (for testing)
	if req.Staging {
		args = append(args, "--server", "letsencrypt_test")
		logCertOp("Using Let's Encrypt STAGING server (test certificates)")
	} else {
		args = append(args, "--server", "letsencrypt")
		logCertOp("Using Let's Encrypt production server")
	}

	// Add challenge type
	if req.Challenge == "http-01" {
		args = append(args, "--webroot", "/var/www/html")
	} else if req.Challenge == "dns-01" {
		// Convert provider name to acme.sh format (usually dns_<provider>)
		dnsProvider := "dns_" + strings.ToLower(req.Provider)
		args = append(args, "--dns", dnsProvider)
	} else if req.Challenge == "tls-alpn-01" {
		args = append(args, "--alpn")
	}

	logCertOp(fmt.Sprintf("Acme.sh command: acme.sh %s", strings.Join(args, " ")))

	// Set environment variables for DNS providers
	cmd := exec.Command("acme.sh", args...)
	env := os.Environ()

	// Add all credentials as environment variables
	if req.Challenge == "dns-01" {
		// Map provider names to acme.sh environment variable names
		envVarMap := map[string]string{
			"duckdns":    "DuckDNS_Token",
			"cloudflare": "CF_Token", // or CF_Key/CF_Email
			"digitalocean": "DO_API_KEY",
			"godaddy":    "GD_Key", // and GD_Secret
			"namecheap":  "NAMECHEAP_API_USER", // and NAMECHEAP_API_KEY
			// Add other providers as needed
		}

		for key, value := range req.Credentials {
			if value != "" {
				// Use mapped env var name if available, otherwise use the key as-is
				envVarName := key
				if mapped, exists := envVarMap[strings.ToLower(req.Provider)]; exists {
					envVarName = mapped
				}
				env = append(env, envVarName+"="+value)
				logCertOp(fmt.Sprintf("Setting environment variable: %s", envVarName))
			}
		}
		logCertOp("DNS-01 challenge selected")
	}

	cmd.Env = env

	// Execute acme.sh with timeout
	logCertOp("Executing acme.sh command...")

	// Create a channel to signal command completion
	done := make(chan error, 1)
	var output []byte

	go func() {
		var err error
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	// Wait for command to complete or timeout (10 minutes for DNS-01, 2 minutes for HTTP-01)
	timeout := 2 * time.Minute
	if req.Challenge == "dns-01" {
		timeout = 10 * time.Minute
		logCertOp("Using 10-minute timeout for DNS-01 challenge")
	} else {
		logCertOp("Using 2-minute timeout for HTTP-01 challenge")
	}

	var err error
	select {
	case err = <-done:
		// Command completed
		logCertOp(fmt.Sprintf("Acme.sh command completed. Output length: %d bytes", len(output)))
	case <-time.After(timeout):
		cmd.Process.Kill()
		err = fmt.Errorf("certificate obtain timeout after %v", timeout)
		logCertOp(fmt.Sprintf("ERROR: %s", err.Error()))
	}

	// Log output
	logCertOp(fmt.Sprintf("Acme.sh output:\n%s", string(output)))

	if err != nil {
		logCertOp(fmt.Sprintf("ERROR: Certificate obtain failed: %v", err))
		sendJSON(w, map[string]interface{}{
			"success": false,
			"output":  string(output),
			"error":   err.Error(),
		})
		return
	}

	// Copy certificates to ssl directory
	// acme.sh stores certs in ~/.acme.sh/domain_ecc/
	// Use the first domain as the primary domain for certificate storage
	primaryDomain := req.Domains[0]

	// For source paths, use the actual domain name (acme.sh keeps wildcard in folder names)
	sourceDomain := primaryDomain

	// For destination paths, remove wildcard prefix to create clean filenames
	destDomain := primaryDomain
	if strings.HasPrefix(destDomain, "*.") {
		destDomain = destDomain[2:]
	}

	domainDir := sourceDomain + "_ecc"
	// acme.sh uses fullchain.cer for the certificate (without domain prefix)
	// but uses domain.key for the key file (with full domain name including wildcard)
	certSource := filepath.Join(acmeDir, domainDir, "fullchain.cer")
	keySource := filepath.Join(acmeDir, domainDir, sourceDomain+".key")

	certDest := filepath.Join(sslDir, destDomain+".crt")
	keyDest := filepath.Join(sslDir, destDomain+".key")

	logCertOp(fmt.Sprintf("Copying certificates from %s to %s", certSource, certDest))

	// Copy cert file
	if data, err := os.ReadFile(certSource); err == nil {
		if err := os.WriteFile(certDest, data, 0644); err != nil {
			logCertOp(fmt.Sprintf("ERROR: Failed to write cert file: %v", err))
		} else {
			logCertOp(fmt.Sprintf("Successfully copied certificate to %s", certDest))
		}
	} else {
		logCertOp(fmt.Sprintf("ERROR: Failed to read cert file: %v", err))
	}

	// Copy key file
	if data, err := os.ReadFile(keySource); err == nil {
		if err := os.WriteFile(keyDest, data, 0600); err != nil {
			logCertOp(fmt.Sprintf("ERROR: Failed to write key file: %v", err))
		} else {
			logCertOp(fmt.Sprintf("Successfully copied key to %s", keyDest))
		}
	} else {
		logCertOp(fmt.Sprintf("ERROR: Failed to read key file: %v", err))
	}

	logCertOp(fmt.Sprintf("Certificate obtain completed successfully for domains: %v\n========================================", req.Domains))

	sendJSON(w, map[string]interface{}{
		"success":  true,
		"output":   string(output),
		"certFile": certDest,
		"keyFile":  keyDest,
	})
}

// Read last N lines of a file
func readLastLines(filePath, lines string) string {
	cmd := exec.Command("tail", "-n", lines, filePath)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("Error reading log: %v", err)
	}
	return string(output)
}

// Helper functions
func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
