package cmd

import (
	"fmt"
	stdlib_html "html"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"codeberg.org/splitringresonator/multiband/docs"
	docs_cli "codeberg.org/splitringresonator/multiband/internal/cli/docs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type mdHandler struct {
	fs       http.FileSystem
	renderer *html.Renderer
	parser   *parser.Parser
}

func newMDRenderer() *html.Renderer {
	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return renderer
}

type anyDirs interface {
	len() int
	name(i int) string
	isDir(i int) bool
}

type fileInfoDirs []fs.FileInfo

func (d fileInfoDirs) len() int          { return len(d) }
func (d fileInfoDirs) isDir(i int) bool  { return d[i].IsDir() }
func (d fileInfoDirs) name(i int) string { return d[i].Name() }

type dirEntryDirs []fs.DirEntry

func (d dirEntryDirs) len() int          { return len(d) }
func (d dirEntryDirs) isDir(i int) bool  { return d[i].IsDir() }
func (d dirEntryDirs) name(i int) string { return d[i].Name() }

func (h *mdHandler) dirList(w http.ResponseWriter, r *http.Request, path string, f http.File) {
	// Prefer to use ReadDir instead of Readdir,
	// because the former doesn't require calling
	// Stat on every entry of a directory on Unix.
	var dirs anyDirs
	var err error
	if d, ok := f.(fs.ReadDirFile); ok {
		var list dirEntryDirs
		list, err = d.ReadDir(-1)
		dirs = list
	} else {
		var list fileInfoDirs
		list, err = f.Readdir(-1)
		dirs = list
	}

	if err != nil {
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs.name(i) < dirs.name(j) })

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<!doctype html>\n")
	fmt.Fprintf(w, "<meta name=\"viewport\" content=\"width=device-width\">\n")
	fmt.Fprintf(w, "<pre>\n")
	if path != "/" {
		fmt.Fprintf(w, "<a href=\"%s\">..</a>\n", filepath.Dir(path))
	}
	for i, n := 0, dirs.len(); i < n; i++ {
		name := dirs.name(i)
		displayName := name
		if dirs.isDir(i) {
			displayName += "/"
		}
		// name may contain '?' or '#', which must be escaped to remain
		// part of the URL path, and not indicate the start of a query
		// string or fragment.
		url := url.URL{Path: filepath.Join(path, name)}
		fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", url.String(), stdlib_html.EscapeString(displayName))
	}
	fmt.Fprintf(w, "</pre>\n")
}

func (h mdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract the filename from the request URL path
	filename := r.URL.Path
	//if filename != "/" && strings.HasSuffix(filename, "/") {
	//	filename = filename[1:] // Remove leading slash
	//}

	file, err := h.fs.Open(filename)
	if os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	d, err := file.Stat()
	if err != nil {
		http.Error(w, "Stat error", http.StatusInternalServerError)
		return
	}

	if d.IsDir() {
		modtime := d.ModTime()
		if !(modtime.IsZero() || modtime.Equal(time.Unix(0, 0))) {
			w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
		}
		h.dirList(w, r, filename, file)
		//h.serveIndex(w, r)
		return
	}

	raw, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Error", http.StatusInternalServerError)
	}

	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	parser := parser.NewWithExtensions(extensions)

	doc := parser.Parse(raw)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "%s", markdown.Render(doc, h.renderer))
}

func (h *mdHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	// List the contents of the directory
	path := r.URL.Path[1:] // Remove leading slash
	dirEntries, err := h.fs.Open(path)
	if os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Could not list directory contents.", http.StatusInternalServerError)
		return
	}
	defer dirEntries.Close()

	// Read the directory entries
	fileInfos, err := dirEntries.Readdir(-1)
	if err != nil {
		http.Error(w, "Could not read directory entries.", http.StatusInternalServerError)
		return
	}

	// Define a template to render the directory listing
	indexTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Index of {{ .Path }}</title>
</head>
<body>
    <h1>Index of {{ .Path }}</h1>
    <ul>
        {{range .Files}}
            <li><a href="{{ .Name }}">{{ .Name }}</a></li>
        {{end}}
    </ul>
</body>
</html>`

	// Parse the template and execute it with the directory path and file info
	t := template.Must(template.New("index").Parse(indexTemplate))
	err = t.ExecuteTemplate(w, "index", map[string]interface{}{
		"Path":  path,
		"Files": getFileInfos(fileInfos),
	})
	if err != nil {
		http.Error(w, "Could not render template.", http.StatusInternalServerError)
	}
}

type fileInfoTemplate struct {
	Name string
}

func getFileInfos(fileInfos []os.FileInfo) []fileInfoTemplate {
	var files []fileInfoTemplate
	for _, fi := range fileInfos {
		if !fi.IsDir() {
			files = append(files, fileInfoTemplate{Name: fi.Name()})
		}
	}
	return files
}

var docsServeCmd = &cobra.Command{
	Use:     "serve",
	GroupID: "docs",
	Short:   "Serve embedded docs over HTTP",
	RunE: func(cmd *cobra.Command, args []string) error {
		host := "127.0.0.1"
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Documentation served at http://%s:%d/\nctrl-c to exit\n", host, port)

		http.Handle("/", mdHandler{fs: http.FS(docs.Docs), renderer: newMDRenderer()})
		return http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
	},
}

var docsCmd = &cobra.Command{
	Use:     "docs",
	GroupID: "docs",
	Short:   "View built-in documentation",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		items := []string{}

		if err := fs.WalkDir(docs.Docs, ".", func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				name := d.Name()

				if name == "." || name == ".." {
					return nil
				}

				return nil // keep walking
			}

			if !d.IsDir() {
				if !strings.HasSuffix(d.Name(), ".md") {
					return nil
				}

				items = append(items, path)
			}
			return nil
		}); err != nil {
			return err
		}

		var width, height uint
		isTerminal := term.IsTerminal(int(os.Stdout.Fd()))
		if isTerminal {
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				width = uint(w)  //nolint:gosec
				height = uint(h) //nolint:gosec
			}

			//if width > 120 {
			//	width = 120
			//}
		}
		//if width == 0 {
		//	width = 80
		//}

		height = 0 // TODO: debug why using reported term height makes pager layout hard to manage with header
		p := tea.NewProgram(docs_cli.NewModel(width, height))

		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	// Local flag: only applies to `serve`.
	docsCmd.AddGroup(&cobra.Group{
		ID:    "docs",
		Title: "Documentation",
	})
	docsServeCmd.Flags().Int("port", 8080, "port to listen on")
	docsCmd.AddCommand(docsServeCmd)
}
