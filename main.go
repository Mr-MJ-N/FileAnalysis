package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type FileItem struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	Size     string `json:"size"`
	TooLarge bool   `json:"too_large"`
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("path")
	if dir == "" {
		dir, _ = os.Getwd()
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var result []FileItem

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		item := FileItem{
			Name:  entry.Name(),
			Path:  NormalizePath(fullPath),
			IsDir: entry.IsDir(),
		}
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				if info.Size() < 10*1024*1024 {
					item.Size = strconv.FormatInt(info.Size()/1024, 10) + " KB"
				} else {
					item.TooLarge = true
				}
			}
		}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir && !result[j].IsDir {
			return true
		} else if !result[i].IsDir && result[j].IsDir {
			return false
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	json.NewEncoder(w).Encode(result)
}

func sizeHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}

	totalSize, err := calcSize(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var sizeStr string
	if totalSize > 1<<30 {
		sizeStr = fmt.Sprintf("%.2f GB", float64(totalSize)/(1<<30))
	} else if totalSize > 1<<20 {
		sizeStr = fmt.Sprintf("%.2f MB", float64(totalSize)/(1<<20))
	} else if totalSize > 1<<10 {
		sizeStr = fmt.Sprintf("%.2f KB", float64(totalSize)/(1<<10))
	} else {
		sizeStr = fmt.Sprintf("%d B", totalSize)
	}

	w.Write([]byte(sizeStr))
}
func calcSize(path string) (int64, error) {
	var total int64 = 0

	err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略无法访问的文件
		}
		if !d.IsDir() {
			info, err := os.Stat(p)
			if err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total, err
}
func NormalizePath(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(indexHTML, "web/index.html"))
	//tmpl := template.Must(template.ParseFiles("web/index.html"))
	tmpl.Execute(w, nil)
}

//go:embed web/index.html
var indexHTML embed.FS

//go:embed web/static/*
var staticFS embed.FS

func staticHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/static/"):] // 去掉前缀
	data, err := staticFS.ReadFile("web/static/" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch {
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css")
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	w.Write(data)
}
func main() {
	port := flag.Int("port", 56541, "服务端口")
	flag.Parse()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/size", sizeHandler)
	http.HandleFunc("/static/", staticHandler)
	dir, _ := os.Getwd()
	log.Printf("打开浏览器访问：http://localhost:%v", *port)
	println("默认目录：" + dir)
	err := http.ListenAndServe(fmt.Sprintf(":%v", *port), nil)
	if err != nil {
		log.Print(err)
		return
	}
}
