// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package main

import (
	"errors"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
	"os"
	"path/filepath"
	"time"
)

type service struct{ data string }

func (s *service) Execute(_ []string, requests <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	status <- svc.Status{State: svc.StartPending}
	stop := make(chan struct{})
	finished := make(chan error, 1)
	go func() { finished <- run(s.data, stop) }()
	status <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
		case err := <-finished:
			if err != nil {
				return true, 1
			}
			return false, 0
		case r := <-requests:
			switch r.Cmd {
			case svc.Interrogate:
				status <- r.CurrentStatus
			case svc.Stop, svc.Shutdown:
				pending := svc.Status{State: svc.StopPending, WaitHint: 30000, CheckPoint: 1}
				status <- pending
				close(stop)
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case err := <-finished:
						if err != nil {
							return true, 1
						}
						return false, 0
					case <-ticker.C:
						pending.CheckPoint++
						status <- pending
					}
				}
			}
		}
	}
}
func runService(data string) error {
	is, err := svc.IsWindowsService()
	if err != nil {
		return err
	}
	if is {
		return svc.Run("SVNova", &service{data})
	}
	return run(data, make(chan struct{}))
}
func serviceCommand(data string, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: svnova --work-path PATH service install|start|stop|uninstall")
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	if args[0] == "install" {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		s, err := m.CreateService("SVNova", exe, mgr.Config{DisplayName: "SVNova", Description: "SVN repository browser and mirror manager", StartType: mgr.StartAutomatic}, "--work-path", data)
		if err != nil {
			return err
		}
		defer s.Close()
		return nil
	}
	s, err := m.OpenService("SVNova")
	if err != nil {
		return err
	}
	defer s.Close()
	switch args[0] {
	case "start":
		return s.Start()
	case "stop":
		_, err = s.Control(svc.Stop)
		return err
	case "uninstall":
		state, err := s.Query()
		if err != nil {
			return err
		}
		if state.State != svc.Stopped {
			return errors.New("stop service before uninstall")
		}
		return s.Delete()
	}
	return errors.New("unknown service command")
}
func lockInstance(data string) (func(), error) {
	p, err := windows.UTF16PtrFromString(filepath.Join(data, "instance.lock"))
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(p, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_ALWAYS, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, errors.New("another instance is running or data directory is inaccessible")
	}
	return func() { windows.CloseHandle(handle) }, nil
}
