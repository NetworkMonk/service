package service

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var elog debug.Log

type fn func()
type myservice struct {
	onRun fn
}

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
			"useage: %s <command>\n"+
			"commands: install, remove, debug, start, stop, pause or continue\n",
		errmsg,
		os.Args[0],
	)
	os.Exit(2)
}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	fasttick := time.Tick(500 * time.Millisecond)
	slowtick := time.Tick(2 * time.Second)
	tick := fasttick
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	elog.Info(1, strings.Join(args, "-"))
	if m.onRun != nil {
		go m.onRun()
	}
loop:
	for {
		select {
		case <-tick:
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				tick = slowtick
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				tick = fasttick
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

// Run will start the main execution loop
func Run(name string, isDebug bool, onRun fn) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &myservice{
		onRun: onRun,
	})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}

// Handle will handle common service request
func Handle(svcName string, svcTitle string, onRun fn) error {
	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("Failed to determine if we are running in an interactive session: %v", err)
	}
	if !isIntSess {
		Run(svcName, false, onRun)
		return nil
	}

	if len(os.Args) < 2 {
		usage("No command specified")
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "debug":
		Run(svcName, true, onRun)
		return nil
	case "install":
		err = Install(svcName, svcTitle)
	case "remove":
		err = Remove(svcName)
	case "start":
		err = Start(svcName)
	case "stop":
		err = Control(svcName, svc.Stop, svc.Stopped)
	case "pause":
		err = Control(svcName, svc.Pause, svc.Paused)
	case "continue":
		err = Control(svcName, svc.Continue, svc.Running)
	default:
		usage(fmt.Sprintf("Invalid command %s", cmd))
	}

	if err != nil {
		log.Fatalf("Failed to %s %s: %v", cmd, svcName, err)
	}

	return nil
}
