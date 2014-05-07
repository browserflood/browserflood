package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

type SimulateArgs struct{
	Config Config
	RunId string
	UserId int
}
type SimulateResult struct {
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
func (s *Slave) SimulateUser(args *SimulateArgs, reply *SimulateResult) error {
	port, err := s.freePort()
	if err != nil {
		return err
	}
	logDir := filepath.Join("log", args.RunId, fmt.Sprintf("%d", args.UserId))
	if err := os.MkdirAll(logDir, 0777); err != nil {
		return err
	}
	phantomLog, err := openLogFile(filepath.Join(logDir, "phantom.log"))
	if err != nil {
		return err
	}
	defer phantomLog.Close()
	phantom := exec.Command("./bin/phantomjs", "--webdriver=127.0.0.1:"+port)
	phantom.Stdout = phantomLog
	phantom.Stderr = phantomLog
	reply.Port = port
	if err := phantom.Start(); err != nil {
		return err
	}
	defer phantom.Process.Kill()
	if err := waitForPort(port); err != nil {
		return err
	}
	phantomExit := make(chan error, 1)
	go func() {
		phantomExit <- phantom.Wait()
	}()
	controlLog, err := openLogFile(filepath.Join(logDir, "control.log"))
	if err != nil {
		return err
	}
	cmdArgs := strings.Split(args.Config.Cmd, " ")
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "PORT="+port)
	cmd.Stdout = controlLog
	cmd.Stderr = controlLog
	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Process.Kill()
	controlExit := make(chan error, 1)
	go func() {
		controlExit <- cmd.Wait()
	}()
	select {
	case err := <-phantomExit:
		return fmt.Errorf("PhantomJS process died: %s", err)
	case err := <-controlExit:
		if err != nil {
			return fmt.Errorf("Control process died: %s", err)
		}
		return nil
	}
}

func openLogFile(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
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

func waitForPort(port string) error {
	start := time.Now()
	for {
		if time.Since(start) > 2*time.Second {
			return fmt.Errorf("Unable to connect to port: %s", port)
		}
		conn, err := net.DialTimeout("tcp", "localhost:"+port, 100*time.Millisecond)
		if err != nil {
			continue
		}
		conn.Close()
		return nil
	}
}
