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
	"github.com/hajimehoshi/oto"
	"github.com/tosone/minimp3"
)

var Version = "version is set by build process"

//embed the assets folder
//go:embed assets
var vfs embed.FS // virtual FileSystem

func main() {
	var up bool
	var errCheck error
	// process flags
	timeout := flag.Duration("timeout", 10*time.Minute, "connection timeout. valid time units are ns, us, ms, s, m, h")
	flag.Parse()
	tmpDir := os.TempDir()

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
	tOut := int(timeout.Seconds())
	log.Printf("Check %s:%s, Timeout: %ds\n", host, port, tOut)

	// progress bar
	bar := pb.Simple.Start(tOut)

	for i := 0; i < tOut; i++ {
		bar.Increment()
		up, errCheck = tcpHealthCheck(host + ":" + port)
		if err != nil {
			log.Println(err)
		}
		if up {
			break
		}
		time.Sleep(1 * time.Second)
	}

	bar.Finish()

	if up {
		log.Printf("%s:%s is up!\n", host, port)
		err := copyFileFromVFS("assets/icons/up.png", tmpDir+"/up.png", vfs)
		if err != nil {
			panic(err)
		}

		err = beeep.Notify("Up!", fmt.Sprintf("%s:%s is Up!", host, port), tmpDir+"/up.png")
		if err != nil {
			panic(err)
		}

	} else {
		log.Printf("%s:%s is down...\n", host, port)
		err := copyFileFromVFS("assets/icons/down.png", tmpDir+"/down.png", vfs)
		if err != nil {
			panic(err)
		}
		err = beeep.Notify("Down!", fmt.Sprintf("%s:%s is Down...(%s)", host, port, errCheck), tmpDir+"/down.png")
		if err != nil {
			panic(err)
		}
	}

	err = copyFileFromVFS("assets/sounds/notification.mp3", tmpDir+"/notification.mp3", vfs)
	if err != nil {
		panic(err)
	}
	err = playSound(tmpDir + "/notification.mp3")
	if err != nil {
		panic(err)
	}
	os.Exit(0)
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

func checkFormat(s string) bool {
	return regexp.MustCompile(`.*:[0-9]+$`).MatchString(s)
}

// tcpHealthCheck will return true if the dest is up, or false if is down. host format should be <host>:<port>
func tcpHealthCheck(host string) (bool, error) {
	conn, err := net.DialTimeout("tcp", host, time.Second)
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

func playSound(sound string) error {
	var file []byte
	file, err := ioutil.ReadFile(sound)
	if err != nil {
		return err
	}

	dec, data, err := minimp3.DecodeFull(file)
	if err != nil {
		return err
	}

	context, err := oto.NewContext(dec.SampleRate, dec.Channels, 2, 1024)
	if err != nil {
		return err
	}

	player := context.NewPlayer()
	player.Write(data)

	<-time.After(time.Second)

	dec.Close()
	if err = player.Close(); err != nil {
		return err
	}
	return nil
}
