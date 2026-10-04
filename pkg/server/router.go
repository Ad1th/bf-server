package server

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/madith/bf-server/pkg/interpreter"
)

// Route represents a mapped Brainfuck route and compiled program.
type Route struct {
	URLPath   string
	FilePath  string
	ModTime   time.Time
	Program   *interpreter.Program
	IsDefault bool // For 404 or 500 fallback handlers
}

// Router manages routing table and Brainfuck program caching / reloading.
type Router struct {
	AppDir     string
	DevMode    bool
	mu         sync.RWMutex
	routes     map[string]*Route
	notFound   *Route
	serverErr  *Route
	lastLoaded time.Time
}

// NewRouter initializes and loads routes from appDir.
func NewRouter(appDir string, devMode bool) (*Router, error) {
	r := &Router{
		AppDir:  appDir,
		DevMode: devMode,
		routes:  make(map[string]*Route),
	}

	if err := r.Reload(); err != nil {
		return nil, err
	}

	return r, nil
}

// Reload scans the app directory and compiles all .bf files.
func (r *Router) Reload() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if directory exists
	info, err := os.Stat(r.AppDir)
	if err != nil {
		return fmt.Errorf("app directory %q does not exist or cannot be accessed: %w", r.AppDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("app path %q is not a directory", r.AppDir)
	}

	newRoutes := make(map[string]*Route)
	var notFoundRoute *Route
	var serverErrRoute *Route

	err = filepath.WalkDir(r.AppDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".bf") {
			return nil
		}

		relPath, err := filepath.Rel(r.AppDir, path)
		if err != nil {
			return err
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		prog, err := interpreter.Compile(string(content))
		if err != nil {
			return fmt.Errorf("compile error in %s: %w", path, err)
		}

		baseName := d.Name()
		// Check for special handlers
		if relPath == "404.bf" || baseName == "404.bf" {
			notFoundRoute = &Route{
				URLPath:   "404",
				FilePath:  path,
				ModTime:   fileInfo.ModTime(),
				Program:   prog,
				IsDefault: true,
			}
		}

		if relPath == "500.bf" || baseName == "500.bf" {
			serverErrRoute = &Route{
				URLPath:   "500",
				FilePath:  path,
				ModTime:   fileInfo.ModTime(),
				Program:   prog,
				IsDefault: true,
			}
		}

		// Convert relative path to URL route
		// e.g. "index.bf" -> "/" and "/index"
		// "hello.bf" -> "/hello"
		// "api/users.bf" -> "/api/users"
		// "api/index.bf" -> "/api" and "/api/index"
		urlPath := "/" + strings.TrimSuffix(filepath.ToSlash(relPath), ".bf")

		route := &Route{
			URLPath:  urlPath,
			FilePath: path,
			ModTime:  fileInfo.ModTime(),
			Program:  prog,
		}

		newRoutes[urlPath] = route

		if strings.HasSuffix(urlPath, "/index") {
			dirURL := strings.TrimSuffix(urlPath, "/index")
			if dirURL == "" {
				dirURL = "/"
			}
			newRoutes[dirURL] = &Route{
				URLPath:  dirURL,
				FilePath: path,
				ModTime:  fileInfo.ModTime(),
				Program:  prog,
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	r.routes = newRoutes
	r.notFound = notFoundRoute
	r.serverErr = serverErrRoute
	r.lastLoaded = time.Now()

	return nil
}

// Match looks up a route for the given request URL path.
// In DevMode, it automatically refreshes if files on disk have changed.
func (r *Router) Match(urlPath string) (*Route, error) {
	cleanPath := pathClean(urlPath)

	if r.DevMode {
		// In dev mode, re-scan and reload routes to pick up changes/edits
		if err := r.Reload(); err != nil {
			return nil, err
		}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Direct match
	if route, exists := r.routes[cleanPath]; exists {
		return route, nil
	}

	// 2. Trailing slash or index fallback
	if cleanPath == "" {
		cleanPath = "/"
	}
	if route, exists := r.routes[cleanPath]; exists {
		return route, nil
	}

	if route, exists := r.routes[cleanPath+"/index"]; exists {
		return route, nil
	}

	// 3. Fallback to custom 404 if defined
	if r.notFound != nil {
		return r.notFound, nil
	}

	return nil, nil
}

// Get500Handler returns the custom 500 handler if configured.
func (r *Router) Get500Handler() *Route {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.serverErr
}

// RouteCount returns the total number of mapped routes.
func (r *Router) RouteCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.routes)
}

// RoutesList returns a list of registered URL paths.
func (r *Router) RoutesList() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	paths := make([]string, 0, len(r.routes))
	for p := range r.routes {
		paths = append(paths, p)
	}
	return paths
}

func pathClean(p string) string {
	if p == "" || p == "/" {
		return "/"
	}
	p = filepath.Clean(p)
	p = filepath.ToSlash(p)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}
