package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	browserFloodPkg  = "github.com/browserflood/browserflood"
	phantomVersion   = "1.9.7"
	phantomDarwinURL = "https://bitbucket.org/ariya/phantomjs/downloads/phantomjs-%s-macosx.zip"
)

func init() {
	register("run", runCmd, "Runs a load test.")
}

func runCmd() error {
	var c int
	var t time.Duration
	set := &flag.FlagSet{}
	set.IntVar(&c, "c", 0, "")
	set.DurationVar(&t, "t", 0, "")
	if err := set.Parse(os.Args[2:]); err != nil {
		return err
	}
	if c == 0 {
		return fmt.Errorf("Missing -c argument.")
	}
	if t == 0 {
		return fmt.Errorf("Missing -t argument.")
	}
	p, err := LoadProject()
	if err != nil {
		return err
	}

	runId := time.Now().Format("2006-02-01-15-04-05")
	logDir := filepath.Join("log", runId)
	latestLogDir := filepath.Join("log", "latest")
	if err := os.MkdirAll(logDir, 0777); err != nil {
		return err
	}
	absLogDir, err := filepath.Abs(logDir)
	if err != nil {
		return err
	}
	if err := os.Symlink(absLogDir, latestLogDir); err != nil {
		return err
	}
	errLog, err := openLogFile(filepath.Join(logDir, "error.log"))
	if err != nil {
		return err
	}
	defer errLog.Close()
	fmt.Printf("Starting run, see \"%s\"\n", logDir)

	targets := map[target]bool{}
	for _, host := range p.Hosts {
		targets[target{Arch: host.Arch, OS: host.OS}] = true
	}
	for target, _ := range targets {
		if err := crossCompileBrowserflood(target); err != nil {
			return err
		}
		if err := downloadPhantomJS(target, phantomVersion); err != nil {
			return err
		}
	}
	fmt.Printf("Syncing files to %d hosts\n", len(p.Hosts))
	results := make(chan error, len(p.Hosts))
	for _, host := range p.Hosts {
		go func() {
			results <- deploy(p, host)
		}()
	}
	for _ = range p.Hosts {
		if err := <-results; err != nil {
			return err
		}
	}
	fmt.Printf("Connecting to %d hosts\n", len(p.Hosts))
	errCh := make(chan error, len(p.Hosts))
	hostCh := make(chan *remoteHost, len(p.Hosts))
	for _, host := range p.Hosts {
		go func() {
			if remote, err := dialHost(p, host); err != nil {
				errCh <- err
			} else {
				hostCh <- remote
			}
		}()
	}
	remotes := make([]*remoteHost, 0, len(p.Hosts))
	for _ = range p.Hosts {
		select {
		case err := <-errCh:
			return err
		case host := <-hostCh:
			remotes = append(remotes, host)
		}
	}
	fmt.Printf("Simulating %d users for %s\n", c, t)
	ids := make(chan int)
	go generateIds(ids)
	for i := 0; i < c; i++ {
		remote := remotes[i%len(remotes)]
		go func() {
			for {
				result, err := remote.SimulateUser(p, <-ids, runId)
				fmt.Printf("result: %#v err: %#v\n", result, err)
			}
		}()
	}
	doneCh := time.After(t)
	start := time.Now()
	for {
		select {
		case <-time.After(time.Second):
			sec := time.Second * time.Duration(int(time.Since(start).Seconds()))
			fmt.Printf("[%s] %d users, %d errors\n", sec, 0, 0)
		case <-doneCh:
			fmt.Printf("Simulation complete\n")
			return nil
		}
	}
	return nil
}

func generateIds(ch chan int) {
	i := 0
	for {
		ch <- i
		i++
	}
}

func dialHost(p *Project, host *Host) (*remoteHost, error) {
	remote := &remoteHost{HeartbeatError: make(chan error, 1)}
	sshAddr := fmt.Sprintf("%s@%s", host.User, host.Host)
	remote.ssh = exec.Command("ssh", sshAddr, "cd "+p.Config.DeployPath+" && ./bin/browserflood slave")
	remote.ssh.Stderr = os.Stderr
	conn, err := newCmdConn(remote.ssh)
	if err != nil {
		return nil, err
	}
	remote.rpc = rpc.NewClient(conn)
	if err := remote.ssh.Start(); err != nil {
		return nil, err
	}
	// Send a heartbeat to verify the connection is working
	if err := remote.Heartbeat(); err != nil {
		return nil, err
	}
	go remote.sendHeartbeats()
	return remote, nil
}

type remoteHost struct {
	HeartbeatError chan error
	rpc            *rpc.Client
	ssh            *exec.Cmd
}

func (r *remoteHost) Heartbeat() error {
	call := r.rpc.Go("Slave.Heartbeat", &struct{}{}, &struct{}{}, nil)
	select {
	case <-call.Done:
		return nil
	case <-time.After(heartbeatTimeout):
		return heartbeatTimeoutError(heartbeatTimeout)
	}
}

func (r *remoteHost) SimulateUser(p *Project, userId int, runId string) (*SimulateResult, error) {
	args := &SimulateArgs{
		Config: p.Config,
		RunId:  runId,
		UserId: userId,
	}
	result := &SimulateResult{}
	err := r.rpc.Call("Slave.SimulateUser", args, result)
	return result, err
}

func (r *remoteHost) sendHeartbeats() {
	for {
		select {
		case <-time.After(time.Second):
			if err := r.Heartbeat(); err != nil {
				r.HeartbeatError <- err
			}
		}
	}
}

func newCmdConn(cmd *exec.Cmd) (*cmdConn, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	return &cmdConn{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}, nil
}

type cmdConn struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (c *cmdConn) Read(buf []byte) (int, error) {
	return c.stdout.Read(buf)
}

func (c *cmdConn) Write(buf []byte) (int, error) {
	return c.stdin.Write(buf)
}

func (c *cmdConn) Close() error {
	c.cmd.Process.Kill()
	return c.cmd.Wait()
}

func deploy(p *Project, host *Host) error {
	dst := fmt.Sprintf("%s@%s:%s", host.User, host.Host, p.Config.DeployPath)
	if err := execCmd("rsync", "-e", "ssh", "-rz", "var", dst); err != nil {
		return err
	}
	bin := fmt.Sprintf("bin/%s_%s/", host.OS, host.Arch)
	return execCmd("rsync", "-e", "ssh", "-rz", bin, dst+"/bin")
}

func execCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s: %s", err, out)
	}
	return err
}

type target struct {
	Arch string
	OS   string
}

func crossCompileBrowserflood(t target) error {
	// @TODO Using browserflood should not require having go installed. But for
	// now this is ok / will allow us to iterate quickly.
	fmt.Printf("Building browserflood for %s/%s\n", t.OS, t.Arch)
	path := fmt.Sprintf("bin/%s_%s/browserflood", t.OS, t.Arch)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	build := exec.Command("go", "build", "-o", path, browserFloodPkg)
	build.Env = append(os.Environ(), "GOOS="+t.OS, "GOARCH="+t.Arch)
	build.Stderr = os.Stderr
	return build.Run()
}

func downloadPhantomJS(t target, version string) error {
	path := fmt.Sprintf("bin/%s_%s/phantomjs", t.OS, t.Arch)
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	fmt.Printf("Downloading phantomjs %s for %s/%s\n", version, t.OS, t.Arch)
	notSupported := fmt.Errorf(
		"Downloading phantomjs is not supported for %s/%s. Please download and place it in %s manually.",
		t.OS,
		t.Arch,
		path,
	)
	switch t.OS {
	case "darwin":
		if t.Arch != "amd64" {
			return notSupported
		}
		url := fmt.Sprintf(phantomDarwinURL, version)
		// We could also do this without the temporary file and directly pipe the
		// download into the zip reader, but for now this makes debugging the code
		// easier.
		dst := filepath.Join("tmp", filepath.Base(url))
		if err := download(url, dst); err != nil {
			return err
		}
		reader, err := zip.OpenReader(dst)
		if err != nil {
			return err
		}
		defer reader.Close()
		extracted := false
		for _, file := range reader.File {
			if strings.HasSuffix(file.Name, "bin/phantomjs") {
				data, err := file.Open()
				if err != nil {
					return err
				}
				defer data.Close()
				dstFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
				if err != nil {
					return err
				}
				defer dstFile.Close()
				if _, err := io.Copy(dstFile, data); err != nil {
					return err
				}
				extracted = true
				break
			}
		}
		if !extracted {
			return fmt.Errorf("Could not find phantomjs in zip file.")
		}
		return nil
	default:
		return notSupported
	}
	return nil
}

func download(url string, dst string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	file, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(file, res.Body); err != nil {
		return err
	}
	return nil
}
