package main

import (
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

const (
	checkInterval = 30 * time.Second // check interval time
)

var (
	cmd      *exec.Cmd
	cmdMutex sync.Mutex
)

func main() {

	args := os.Args
	if len(args) <= 3 {
		log.Fatal("run server, miss args, please check")
	}

	sn := args[1] // server name
	sb := args[2] // server binary path
	sp := args[3] // server port

	for {
		if !isRunning(sp) {
			log.Printf("%s is not running, start it now", sn)
			err := startService(sn, sb)
			if err != nil {
				log.Printf("start server: [%s] fail: [%v]", sn, err)
			} else {
				log.Printf("start server: [%s] success", sn)
			}
		}
		time.Sleep(checkInterval)
	}

	select {}
}

func startService(sn, sb string) error {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	if cmd != nil && cmd.Process != nil {
		_ = stopService(sn)
	}

	// for ready to cmd
	cmd = exec.Command(sb)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// run server

	err := cmd.Start()
	if err != nil {
		log.Printf("server: [%s] stat run fail\n", sn)
		return err
	}
	log.Printf("server: [%s] stat run success, PID: [%d]\n", sn, cmd.Process.Pid)
	return nil
}

func stopService(sn string) error {
	if cmd == nil || cmd.Process == nil {
		log.Println("no server is running, no need to stop")
		return nil
	}

	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM) // kill process
	if err != nil {
		log.Printf("stop server: [%s] failed, err: %v", sn, err)
		return err
	}

	log.Printf("stop server: [%s] success", sn)
	return nil
}

func isRunning(port string) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", port), 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
