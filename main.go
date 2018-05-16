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

// Because no one ever fucking provides this
func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}
func createEnv() (Env, error) {
	host := os.Getenv("GONAD_HOST")
	port := os.Getenv("GONAD_PORT")
	if port == "" {
		port = "601"
	}
	dest_type := os.Getenv("GONAD_DESTINATION")
	if dest_type == "" {
		dest_type = "stdout"
	}
	dest_path := os.Getenv("GONAD_DESTINATION_PATH")
	if dest_path == "" {
		dest_path = "/var/log/gonad.log"
	}
	if pathExists(path.Dir(dest_path)) == false {
		log.WithFields(logrus.Fields{"path": dest_path}).Error("directory does not exist!")
		os.Exit(1)
	}
	var f *os.File
	var err error
	if dest_type == "file" {
		f, err = os.OpenFile(dest_path, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.WithFields(logrus.Fields{
				"path": dest_path,
			}).Error(err)
			os.Exit(1)
		}
	} else if dest_type == "stdout" {
		f = os.Stdout
	} else if dest_type == "stderr" {
		f = os.Stderr
	}
	return Env{host, port, dest_type, dest_path, f}, nil
}
func main() {
	log.Out = os.Stdout
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
	}).Info("Gonad " + Version + " running")
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
	total_bytes := 0
	br := bufio.NewReader(conn)
	defer func() {
		conn.Close()
		env.Fd.Write([]byte("\n"))
	}()
	for {
		conn.SetReadDeadline(time.Now().Add(maxReadTimeout))
		bytes, err := br.ReadBytes('\n')
		total_bytes += len(bytes)
		if err != nil {
			rep := log.WithFields(logrus.Fields{
				"method":           "handleAccept",
				"while":            "reading from client",
				"remote":           conn.RemoteAddr(),
				"local":            conn.LocalAddr(),
				"total_bytes_read": total_bytes,
			})
			if err == io.EOF {
				rep.Info(err)
			} else {
				rep.Error(err)
			}
			return
		}
		env.Fd.Write(bytes)
		env.Fd.Write([]byte("\n"))
	}
}
