package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Endpoint struct {
	Host string
	Port int
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type SlowPipe struct {
	Local  *Endpoint
	Remote *Endpoint
	Rate   int
}

func (slowPipe *SlowPipe) Start() error {
	listener, err := net.Listen("tcp", slowPipe.Local.String())
	if err != nil {
		fmt.Printf("Failed %s\n", err)
		return err
	}
	defer listener.Close()
	fmt.Printf("Established Listener\n")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %s\n", err)
			return err
		}

		go slowPipe.forward(conn)
	}
}

func (slowPipe *SlowPipe) forward(local net.Conn) {
	remote, err := net.Dial("tcp", slowPipe.Remote.String())
	if err != nil {
		fmt.Printf("Target dial error: %s\n", err)
		return
	}

	copyConn := func(writer, reader net.Conn, label string) {
		bytes, err := io.Copy(writer, reader)
		if err != nil {
			fmt.Printf("Data transfer error %s: %s", label, err)
			return
		}
		fmt.Printf("Data transfer %s %d bytes\n", label, bytes)
	}

	go copyConn(local, remote, "\u2b06")
	go copyConn(remote, local, "\u2b07")
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	// colorize(color.FgGreen, "=> Reading Config")

	const (
		defaultListenPort       = 9090
		defaultListenPortUsage  = "default listen port"
		defaultTargetPort       = 9091
		defaultTargetPortUsage  = "default target port"
		defaultbytesPerSec      = 4
		defaultBytesPerSecUsage = "bytes per second"
	)

	listenPort := flag.Int("listenPort", defaultListenPort, defaultListenPortUsage)
	targetPort := flag.Int("targetPort", defaultTargetPort, defaultTargetPortUsage)
	bytesPerSec := flag.Int("bytesPerSec", defaultbytesPerSec, defaultBytesPerSecUsage)

	flag.Parse()

	// colorize(color.FgCyan, "Listen: ", *listenPort)
	// colorize(color.FgCyan, "Target: ", *targetPort)
	// colorize(color.FgCyan, "Rate: ", *bytesPerSec)

	slowPipe := &SlowPipe{
		Local:  &Endpoint{Host: "localhost", Port: *listenPort},
		Remote: &Endpoint{Host: "localhost", Port: *targetPort},
		Rate:   *bytesPerSec,
	}

	go slowPipe.Start()

	// Wait for Ctrl-C
	<-done
	fmt.Println("Exiting")
}
