package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"time"
)

const heartbeatTimeout = 3 * time.Second

type heartbeatTimeoutError time.Duration

func (h heartbeatTimeoutError) Error() string {
	return fmt.Sprintf("Heartbeat timeout after %s.", h)
}

func slaveCmd() error {
	slave := &Slave{heartbeat: make(chan struct{})}
	go slave.checkHeartbeats()
	s := rpc.NewServer()
	s.Register(slave)
	s.ServeConn(&stdioConn{})
	return nil
}

type stdioConn struct{}

func (s *stdioConn) Read(buf []byte) (int, error) {
	return os.Stdin.Read(buf)
}

func (s *stdioConn) Write(buf []byte) (int, error) {
	return os.Stdout.Write(buf)
}

func (s *stdioConn) Close() error {
	return os.Stdin.Close()
}

type RunArgs struct{}
type RunResult struct {
	Port string
}

type Slave struct {
	heartbeat chan struct{}
}

func (s *Slave) checkHeartbeats() {
	for {
		select {
		case <-s.heartbeat:
			continue
		case <-time.After(heartbeatTimeout):
			fatal("%s", heartbeatTimeoutError(heartbeatTimeout))
		}
	}
}

func (s *Slave) Heartbeat(_ *struct{}, _ *struct{}) error {
	s.heartbeat <- struct{}{}
	return nil
}

// Run runs a single user.
func (s *Slave) Run(args *RunArgs, reply *RunResult) error {
	port, err := s.freePort()
	if err != nil {
		return err
	}
	phantom := exec.Command("./phantomjs", "--webdriver=127.0.0.1:"+port)
	phantom.Stdout = os.Stderr
	phantom.Stderr = os.Stderr
	reply.Port = port
	return phantom.Run()
}

// @TODO this might be prone to race conditions.
func (s *Slave) freePort() (string, error) {
	conn, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_, port, err := net.SplitHostPort(conn.Addr().String())
	return port, err
}
