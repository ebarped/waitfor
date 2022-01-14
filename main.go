package main

import (
	"embed"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/gen2brain/beeep"
)

// default timeout 600s (10 mins)
const timeout int = 600

// tcpHealthCheck will return true if the dest is up, or false if is down. host format should be <host>:<port>
func tcpHealthCheck(host string) (bool, error) {
	conn, err := net.DialTimeout("tcp", host, time.Second*time.Duration(timeout))
	if err != nil {
		return false, fmt.Errorf("failed to connect: %s", err)
	}
	if conn != nil {
		defer conn.Close()
	}
	return true, nil
}

func copyFileFromVFS(src, dest string, vfs embed.FS) error {
	f, err := vfs.ReadFile(src)
	if err != nil {
		return fmt.Errorf("error opening file from embedded fs: %v", err)
	}
	err = ioutil.WriteFile(dest, f, 0o644)
	if err != nil {
		return fmt.Errorf("error writing file from embedded fs: %v", err)
	}
	return nil
}

func checkFormat(s string) bool {
	return regexp.MustCompile(`.*:[0-9]+$`).MatchString(s)
}

// parseRawURL takes an url in any of the following forms:
// - scheme://host:port
// - host:port
// and returns the host and the port fields
func parseRawURL(rawurl string) (host, port string, err error) {
	u, err := url.ParseRequestURI(rawurl) // format scheme://host:port
	if err != nil || u.Host == "" {       // format host:port
		u, repErr := url.ParseRequestURI("https://" + rawurl)
		if repErr != nil {
			return "", "", err
		}
		host = u.Hostname()
		port := u.Port()
		err = nil
		return host, port, nil
	}

	host = u.Hostname()
	port = u.Port()

	return host, port, nil
}

//embed the icons folder
//go:embed icons
var vfs embed.FS // virtual FileSystem

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatalf("error: you should pass <host>:<port> or <ip>:<port> as argument\n")
	}
	host, port, err := parseRawURL(args[1]) // we parse it to get rid of things like the scheme
	if err != nil {
		log.Fatalf("Could not parse raw url: %s, error: %v", args[1], err)
	}

	validFormat := checkFormat(host + ":" + port)
	if !validFormat {
		log.Fatalf("error: the format of the host must be <host>:<port> or <ip>:<port>\n")
	}

	// we save this error to print on the notification
	up, errCheck := tcpHealthCheck(host + ":" + port)
	if errCheck != nil {
		log.Printf("error checking %s:%s: %v\n", host, port, err)
	}

	if up {
		log.Printf("%s:%s is up!\n", host, port)
		err := copyFileFromVFS("icons/up.png", os.TempDir()+"/up.png", vfs)
		if err != nil {
			panic(err)
		}

		err = beeep.Notify("Up!", fmt.Sprintf("%s:%s is Up!", host, port), os.TempDir()+"/up.png")
		if err != nil {
			panic(err)
		}
	} else {
		log.Printf("%s:%s is down...\n", host, port)
		err := copyFileFromVFS("icons/down.png", os.TempDir()+"/down.png", vfs)
		if err != nil {
			panic(err)
		}
		err = beeep.Notify("Down!", fmt.Sprintf("%s:%s is Down...(%s)", host, port, errCheck), os.TempDir()+"/down.png")
		if err != nil {
			panic(err)
		}
	}
}
