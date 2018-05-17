package main

import (
	"bufio"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	path "path/filepath"
	"time"
	"io"
)

var log = logrus.New()

const Version = "1.0.0"

type Env struct {
	Host            string
	Port            string
	Destination     string // stdout, stderr, file, server
	DestinationPath string
	Fd              *os.File
}

// Because no one ever provides this for reasons
func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}
func createEnv() (Env, error) {
	host := os.Getenv("GONNAD_HOST")
	port := os.Getenv("GONNAD_PORT")
	if port == "" {
		port = "601"
	}
	dest_type := os.Getenv("GONNAD_DESTINATION")
	if dest_type == "" {
		dest_type = "stdout"
	}
	dest_path := os.Getenv("GONNAD_DESTINATION_PATH")
	if dest_path == "" {
		dest_path = "/var/log/GONNAD.log"
	}
	if pathExists(path.Dir(dest_path)) == false {
		log.WithFields(logrus.Fields{"path": dest_path}).Error("directory does not exist!")
		os.Exit(1)
	}
	var f *os.File
	if dest_type == "file" {
		f = nil
	} else if dest_type == "stdout" {
		f = os.Stdout
	} else if dest_type == "stderr" {
		f = os.Stderr
	}
	return Env{host, port, dest_type, dest_path, f}, nil
}
func destroyEnv(env Env) {
	if (env.Destination == "file") {
		env.Fd.Close()
	}
	// Here for backend server code
}
func setOutput(f *os.File) {
	logrus.SetOutput(f)
}
func main() {
	setOutput(os.Stdout)
	env, err := createEnv()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	defer env.Fd.Close()
	listener, err := net.Listen("tcp", env.Host+":"+env.Port)
	if err != nil {
		log.WithFields(logrus.Fields{
			"host": env.Host,
			"port": env.Port,
		}).Error(err)
		os.Exit(1)
	}
	defer listener.Close()
	log.WithFields(logrus.Fields{
		"host": env.Host,
		"port": env.Port,
	}).Info("GONNAD " + Version + " running")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		go handleAccept(env, conn, 5*time.Second)
	}
}

func handleAccept(env Env, conn net.Conn, maxReadTimeout time.Duration) {
	var dst *bufio.Writer
	var fd *os.File
	var err error
	total_bytes := 0
	br := bufio.NewReader(conn)
	fd = env.Fd
	defer func(){
		dst.Write([]byte("\n"))
		dst.Flush()
		conn.Close()
		if env.Destination == "file" {
			fd.Close()
		}
	}()
	if env.Destination == "file" {
		fd,err = os.OpenFile(env.DestinationPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.WithFields(logrus.Fields{
				"method":           "handleAccept",
				"while":            "reading from client",
				"remote":           conn.RemoteAddr(),
				"local":            conn.LocalAddr(),
				"total_bytes_read": total_bytes,
			}).Error(err)
			return
		}
	}
	dst = bufio.NewWriter(fd)
	for {
		conn.SetReadDeadline(time.Now().Add(maxReadTimeout))
		bytes, err := br.ReadBytes('\n')
		total_bytes += len(bytes)
		if err != nil {
			dst.Flush()
			rep := log.WithFields(logrus.Fields{
				"method":           "handleAccept",
				"while":            "reading from client",
				"remote":           conn.RemoteAddr(),
				"local":            conn.LocalAddr(),
				"total_bytes_read": total_bytes,
			})
			if err == io.EOF {
				rep.Info(err)
			} else if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
				rep.Info("Timeout on socket " + err.Error())
			} else if oErr, ok := err.(*net.OpError); ok {
				if oErr.Op == "read" {
					rep.Info("could not read due to OpError " + oErr.Error())
				} else {
					rep.Error(err)
				}
			} else {
				rep.Error(err)
			}
			return
		}
		log.Info(string(bytes))
		dst.Write(bytes)
		dst.Write([]byte("\n"))
	}
}
