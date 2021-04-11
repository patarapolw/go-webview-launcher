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
	"net"
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
		for _, env := range cfg.Cmd.Env {
			kv := strings.Split(env, "=")
			v := strings.Join(kv[1:], "=")
			os.Setenv(kv[0], v)

			if kv[0] == "PORT" {
				port = v
			}
		}

		if port == "0" {
			listener, err := net.Listen("tcp", ":0")
			if err != nil {
				panic(err)
			}

			port = strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)

			err = listener.Close()
			if err != nil {
				panic(err)
			}
		}

		cfg.Cmd.Env = append(cfg.Cmd.Env, "PORT="+port)

		e := cfg.Cmd.Start()
		if e != nil {
			panic(e)
		}

		for _, env := range cfg.Cmd.Env {
			kv := strings.Split(env, "=")
			os.Setenv(kv[0], strings.Join(kv[1:], "="))
		}

		defer cfg.Cmd.Process.Kill()
		defer func() {
			if cfg.Kill.Path != "" {
				cfg.Kill.Env = append(cfg.Kill.Env, "PORT="+port)

				e := cfg.Kill.Run()
				if e != nil {
					panic(e)
				}
			}
		}()
	}

	if cfg.URL == "" {
		cfg.URL = fmt.Sprintf("http://localhost:%s", port)
	}

	w := webview.New(cfg.Debug)
	defer w.Destroy()

	w.SetSize(int(C.display_width()), int(C.display_height()), webview.HintNone)
	w.Navigate(cfg.URL)
	w.Run()
}
