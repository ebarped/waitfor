package main

import (
	"embed"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/gen2brain/beeep"
)

// default timeout 600s (10 mins)
const timeout int = 600

// tcpHealthCheck will return true if the dest is up, or false if is down
func tcpHealthCheck(host string) (bool, error) {
	u, err := url.Parse(host)
	if err != nil {
		return false, fmt.Errorf("failed to parse url: %s", err)
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(u.Hostname(), u.Port()), time.Second*time.Duration(timeout))
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

//embed the icons folder
//go:embed icons
var vfs embed.FS // virtual FileSystem

func main() {
	host := "google.es"
	port := "80"

	fmt.Printf("Trying %s:%s...\n", host, port)
	up, err := tcpHealthCheck("tcp://" + host + ":" + port)
	if err != nil {
		log.Printf("error checking %s:%s: %v\n", host, port, err)
	}
	fmt.Printf("Host %s:%s is up? %t\n", host, port, up)

	if up {
		err := copyFileFromVFS("icons/up.png", os.TempDir()+"/up.png", vfs)
		if err != nil {
			panic(err)
		}

		err = beeep.Notify("Up!", fmt.Sprintf("%s:%s is Up!", host, port), os.TempDir()+"/up.png")
		if err != nil {
			panic(err)
		}
	} else {
		err := copyFileFromVFS("icons/down.png", os.TempDir()+"/down.png", vfs)
		if err != nil {
			panic(err)
		}
		err = beeep.Notify("Down!", fmt.Sprintf("%s:%s is Down!", host, port), os.TempDir()+"/down.png")
		if err != nil {
			panic(err)
		}
	}
}
