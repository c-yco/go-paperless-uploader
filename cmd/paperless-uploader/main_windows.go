//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "PaperlessUploader"

var elog debug.Log

func main() {
	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}

	if !isInteractive {
		runService(false)
		return
	}

	if len(os.Args) > 1 {
		cmd := strings.ToLower(os.Args[1])
		switch cmd {
		case "debug":
			runService(true)
			return
		case "install":
			err = installService()
		case "remove":
			err = removeService()
		case "start":
			err = startService()
		case "stop":
			err = controlService(svc.Stop, svc.Stopped)
		case "pause":
			err = controlService(svc.Pause, svc.Paused)
		case "continue":
			err = controlService(svc.Continue, svc.Running)
		default:
			if err := runApp(); err != nil {
				log.Fatalf("Error: %v", err)
			}
			return
		}
		if err != nil {
			log.Fatalf("failed to %s service: %v", cmd, err)
		}
		return
	}

	// if we are not running as a service, and no command was given, run the app
	if err := runApp(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

type paperlessUploaderService struct{}

func (s *paperlessUploaderService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	elog.Info(1, "Paperless Uploader service starting.")
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	go func() {
		if err := runApp(); err != nil {
			elog.Error(1, fmt.Sprintf("runApp failed: %v", err))
		}
	}()

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				elog.Info(1, "Paperless Uploader service stopping.")
				changes <- svc.Status{State: svc.StopPending}
				return
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
}

func runService(isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(serviceName)
	} else {
		elog, err = eventlog.Open(serviceName)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, "Paperless Uploader service starting.")
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(serviceName, &paperlessUploaderService{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("service run failed: %v", err))
		return
	}
	elog.Info(1, "Paperless Uploader service stopped.")
}

func getServiceManager() (*mgr.Mgr, error) {
	return mgr.Connect()
}

func installService() error {
	m, err := getServiceManager()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	exepath, err := os.Executable()
	if err != nil {
		return err
	}

	s, err := m.CreateService(serviceName, exepath, mgr.Config{DisplayName: "Paperless Uploader Service"})
	if err != nil {
		return err
	}
	defer s.Close()

	return nil
}

func removeService() error {
	m, err := getServiceManager()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed", serviceName)
	}
	defer s.Close()

	return s.Delete()
}

func startService() error {
	m, err := getServiceManager()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()

	return s.Start()
}

func controlService(c svc.Cmd, to svc.State) error {
	m, err := getServiceManager()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()

	status, err := s.Control(c)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}

	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if time.Now().After(timeout) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}
