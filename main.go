package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const version = "1.0.0"

func main() {
	host := flag.String("host", "localhost", "host/interface to listen on")
	port := flag.Int("port", 8090, "port to listen on")
	folder := flag.String("folder", "", "video folder to serve (defaults to current directory)")
	nobrowser := flag.Bool("no-browser", false, "do not open a browser automatically")
	verFlag := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "RockTube %s — a tiny self-hosted YouTube-style video server\n\n", version)
		fmt.Fprintf(out, "Drop the binary in a folder of videos and run it. That's it.\n\n")
		fmt.Fprintf(out, "Usage:\n")
		fmt.Fprintf(out, "  rocktube [flags] [--folder PATH]\n\n")
		fmt.Fprintf(out, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(out, "\nExamples:\n")
		fmt.Fprintf(out, "  rocktube                       # serve videos in the current folder\n")
		fmt.Fprintf(out, "  rocktube --folder \"D:\\Movies\"  # serve a specific folder\n")
		fmt.Fprintf(out, "  rocktube --port 9000           # use a different port\n")
	}
	flag.Parse()

	if *verFlag {
		fmt.Println("rocktube", version)
		return
	}

	videoDir, err := resolveFolder(*folder)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	srv := NewServer(videoDir)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	ln, err := net.Listen("tcp", addr)
	fallbackUsed := false
	if err != nil {
		// If the chosen port is taken and the user picked the default, try a few alternatives.
		if *port == 8090 {
			for _, p := range []int{8091, 8092, 8093, 8094, 8095} {
				if l2, e := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, p)); e == nil {
					ln = l2
					*port = p
					fallbackUsed = true
					break
				}
			}
		}
		if ln == nil {
			log.Fatalf("could not listen on %s: %v", addr, err)
		}
	}

	url := fmt.Sprintf("http://%s:%d", preferredHost(*host), *port)
	fmt.Printf("RockTube %s\n", version)
	fmt.Printf("Serving videos from: %s\n", videoDir)
	fmt.Printf("Listening on:        %s\n", url)
	if fallbackUsed {
		fmt.Printf("(default port 8090 was busy — using %d)\n", *port)
	}
	fmt.Println("Press Ctrl+C to stop.")

	if !*nobrowser {
		go openBrowser(url)
	}

	if err := http.Serve(ln, srv.router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func resolveFolder(f string) (string, error) {
	if f == "" {
		f = "."
	}
	abs, err := filepath.Abs(f)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("cannot access folder %q: %w", abs, err)
	}
	if !st.IsDir() {
		return "", fmt.Errorf("%q is not a directory", abs)
	}
	return abs, nil
}

func preferredHost(h string) string {
	if h == "" || h == "0.0.0.0" || h == "::" {
		return "localhost"
	}
	return h
}

func openBrowser(url string) {
	// Give the server a beat to start accepting connections.
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if cmd != nil {
		_ = cmd.Start()
	}
}
