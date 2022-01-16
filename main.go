package main

import (
	"embed"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/gen2brain/beeep"
)

// tcpHealthCheck will return true if the dest is up, or false if is down. host format should be <host>:<port>
func tcpHealthCheck(host string, timeout time.Duration) (bool, error) {
	conn, err := net.Dial("tcp", host)
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
	var up bool
	var errCheck error
	// process flags
	timeout := flag.Duration("timeout", 10*time.Minute, "connection timeout. valid time units are ns, us, ms, s, m, h")
	flag.Parse()

	// process args
	args := flag.Args() // this strips out the flags from the arguments of the program
	if len(args) != 1 {
		fmt.Printf("%#v\n", args)
		log.Fatalf("error: you should pass <host>:<port> or <ip>:<port> as argument\n")
	}

	host, port, err := parseRawURL(args[0]) // we parse it to get rid of things like the scheme
	if err != nil {
		log.Fatalf("Could not parse raw url: %s, error: %v", args[0], err)
	}

	validFormat := checkFormat(host + ":" + port)
	if !validFormat {
		log.Fatalf("error: the format of the host must be <host>:<port> or <ip>:<port>\n")
	}

	log.Printf("Check %s:%s, Timeout: %s\n", host, port, *timeout)

	// progress bar
	bar := pb.Simple.Start(int(timeout.Seconds()))

	for i := 0; i < int(timeout.Seconds()); i++ {
		up, errCheck = tcpHealthCheck(host+":"+port, *timeout)
		time.Sleep(1 * time.Second)
		bar.Increment()
		if up {
			break
		}
	}
	bar.Finish()

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
