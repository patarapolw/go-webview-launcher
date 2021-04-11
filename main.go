package main

/*
#cgo darwin LDFLAGS: -framework CoreGraphics
#cgo linux pkg-config: x11
#if defined(__APPLE__)
#include <CoreGraphics/CGDisplayConfiguration.h>
int display_width() {
	return CGDisplayPixelsWide(CGMainDisplayID());
}
int display_height() {
	return CGDisplayPixelsHigh(CGMainDisplayID());
}
#elif defined(_WIN32)
#include <wtypes.h>
int display_width() {
	RECT desktop;
	const HWND hDesktop = GetDesktopWindow();
	GetWindowRect(hDesktop, &desktop);
	return desktop.right;
}
int display_height() {
	RECT desktop;
	const HWND hDesktop = GetDesktopWindow();
	GetWindowRect(hDesktop, &desktop);
	return desktop.bottom;
}
#else
#include <X11/Xlib.h>
int display_width() {
	Display* d = XOpenDisplay(NULL);
	Screen*  s = DefaultScreenOfDisplay(d);
	return s->width;
}
int display_height() {
	Display* d = XOpenDisplay(NULL);
	Screen*  s = DefaultScreenOfDisplay(d);
	return s->height;
}
#endif
*/
import "C"
import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/webview/webview"
	"gopkg.in/yaml.v2"
)

type config struct {
	Cmd   exec.Cmd
	Kill  exec.Cmd
	URL   string
	Title string
	Dir   string
	Debug bool
}

func main() {
	d, e := ioutil.ReadFile("webview.yaml")
	if e != nil {
		panic(e)
	}

	cfg := config{}

	e = yaml.Unmarshal(d, &cfg)
	if e != nil {
		panic(e)
	}

	port := "0"

	if cfg.Cmd.Path != "" {
		cmd := exec.Command(cfg.Cmd.Path, cfg.Cmd.Args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = cfg.Cmd.Env

		for _, env := range cmd.Env {
			kv := strings.Split(env, "=")
			v := strings.Join(kv[1:], "=")
			os.Setenv(kv[0], v)

			if kv[0] == "PORT" {
				port = v
			}
		}

		if port == "0" {
			port = getRandomPort()
		}

		cmd.Env = append(cmd.Env, "PORT="+port)

		e := cmd.Start()
		if e != nil {
			panic(e)
		}

		defer cmd.Process.Kill()
		defer func() {
			if cfg.Kill.Path != "" {
				cmd := exec.Command(cfg.Kill.Path, cfg.Kill.Args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Env = cfg.Kill.Env

				cmd.Env = append(cmd.Env, "PORT="+port)

				e := cmd.Run()
				if e != nil {
					panic(e)
				}
			}
		}()
	} else if cfg.URL == "" && cfg.Dir != "" {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()
		go http.Serve(listener, http.FileServer(http.Dir(cfg.Dir)))

		port = strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	}

	if cfg.URL == "" {
		cfg.URL = fmt.Sprintf("http://localhost:%s", port)
	}

	fmt.Println("Opening: " + cfg.URL)

	for {
		_, err := http.Get(cfg.URL)
		if err == nil {
			break
		}
	}

	w := webview.New(cfg.Debug)
	defer w.Destroy()

	w.SetSize(int(C.display_width()), int(C.display_height()), webview.HintNone)

	if cfg.Title != "" {
		w.SetTitle(cfg.Title)
	}

	w.Navigate(cfg.URL)
	w.Run()
}

func getRandomPort() string {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)

	err = listener.Close()
	if err != nil {
		panic(err)
	}

	return port
}
