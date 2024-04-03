package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	htmlPage = `<html>
<head>
    <title>Everything</title>
</head>
<body>
    %s
</body>
</html>`
)

// GetWindowsDrives 返回 Windows 系统中的所有盘符列表
func GetWindowsDrives() []string {
	var drives []string
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		drivePath := string(drive) + ":\\"
		_, err := os.Stat(drivePath)
		if err == nil {
			drives = append(drives, drivePath)
		}
	}
	return drives
}

// ListDirectories 返回指定路径下的所有一级文件夹列表
func ListDirectories(path string) ([]string, error) {
	var dirs []string
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}
	return dirs, nil
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	drives := GetWindowsDrives()
	var html strings.Builder
	html.WriteString("<ul>")
	for _, drive := range drives {
		html.WriteString(fmt.Sprintf(`<li><a href="/%s">%s</a></li>`, drive, drive))
	}
	html.WriteString("</ul>")
	fmt.Fprintf(w, htmlPage, html.String())
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 解析请求路径
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	drive := parts[1]
	if drive == "" {
		// 如果路径中没有盘符，则返回根目录
		rootHandler(w, r)
		return
	}

	path := strings.Join(parts[2:], "/")
	if path == "" {
		// 如果路径中只包含盘符，则返回盘符下的所有文件和文件夹
		driveHandler(w, r, drive)
		return
	}
	if drive == "download" {
		http.ServeFile(w, r, path)
		return
	}

	// 处理带有路径的请求
	filePath := strings.Join(parts[1:], "/") //filepath.Join(drive, path)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if fileInfo.IsDir() {
		// 如果请求的是文件夹，则列出文件夹下的内容
		directoryHandler(w, r, filePath)
		return
	}

	// 如果请求的是文件，则直接下载文件
	http.ServeFile(w, r, filePath)
}

func driveHandler(w http.ResponseWriter, r *http.Request, drive string) {
	path := filepath.Join(drive, string(filepath.Separator))
	dirs, err := ListDirectories(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	var html strings.Builder
	html.WriteString("<ul>")
	for _, dir := range dirs {
		html.WriteString(fmt.Sprintf(`<li><a href="/%s/%s">%s</a></li>`, drive, dir, dir))
	}
	html.WriteString("</ul>")
	fmt.Fprintf(w, htmlPage, html.String())
}

func directoryHandler(w http.ResponseWriter, r *http.Request, path string) {
	dirs, err := ListDirectories(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	var html strings.Builder
	html.WriteString("<ul>")
	for _, dir := range dirs {
		html.WriteString(fmt.Sprintf(`<li><a href="/%s">%s</a></li>`, filepath.Join(path, dir), dir))
	}
	for _, file := range files {
		if !file.IsDir() {
			html.WriteString(fmt.Sprintf(`<li><a href="/download/%s">%s</a></li>`, filepath.Join(path, file.Name()), file.Name()))
		}
	}
	html.WriteString("</ul>")
	fmt.Fprintf(w, htmlPage, html.String())
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest)
	mux.HandleFunc("/download/*", handleRequest)
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// Do nothing for favicon requests
	})

	srv := &http.Server{
		Addr:    ":1688",
		Handler: mux,
	}

	fmt.Println("dir server run at 1688 port")
	fmt.Println("group: https://t.me/hackse0, telegram: @ty1904")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
