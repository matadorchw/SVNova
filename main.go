// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package main

import (
	"context"
	"embed"
	"flag"

	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"svnova/internal/bundle"
	"svnova/services/app"
)

// Include underscore-prefixed shared chunks emitted by Vite.
//
//go:embed templates all:public/assets
var resources embed.FS

var installListen = "127.0.0.1:3000"

func main() {
	executable, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	data := flag.String("work-path", filepath.Dir(executable), "application work directory containing custom, data and log")
	flag.StringVar(&installListen, "install-listen", installListen, "installation listener for an unconfigured instance")
	flag.Parse()
	absolute, err := filepath.Abs(*data)
	if err != nil {
		log.Fatal(err)
	}
	if flag.NArg() > 0 && flag.Arg(0) == "service" {
		if err = serviceCommand(absolute, flag.Args()[1:]); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err = runService(absolute); err != nil {
		log.Fatal(err)
	}
}
func run(data string, stop <-chan struct{}) error {
	if err := app.PrepareWorkPath(data); err != nil {
		return err
	}
	unlock, err := lockInstance(data)
	if err != nil {
		return err
	}
	defer unlock()
	logFile, err := os.OpenFile(filepath.Join(data, "log", "svnova.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer logFile.Close()
	previousLog := log.Writer()
	log.SetOutput(io.MultiWriter(previousLog, logFile))
	defer log.SetOutput(previousLog)
	bin, err := bundle.Extract(filepath.Join(data, "data"))
	if err != nil {
		return err
	}
	templates, err := app.ParseTemplates(resources)
	if err != nil {
		return err
	}
	assets, err := fs.Sub(resources, "public")
	if err != nil {
		return err
	}
	application, err := app.New(data, bin, templates, http.FileServer(http.FS(assets)))
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	go func() {
		select {
		case <-stop:
			cancel()
		case <-ctx.Done():
		}
	}()
	address := application.Config.Listen
	if address == "" {
		address = installListen
	}
	server := &http.Server{Addr: address, Handler: application.Handler(), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	log.Printf("SVNova listening on %s; work path: %s", address, data)
	err = serveUntilStopped(ctx, server, listener, application.Background, 15*time.Second)
	if application.DB != nil {
		if closeErr := application.DB.Close(); err == nil {
			err = closeErr
		}
	}
	return err
}

// Keep the instance lock until requests and repository processes have stopped.
func serveUntilStopped(parent context.Context, server *http.Server, listener net.Listener, background func(context.Context), grace time.Duration) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	var requests sync.WaitGroup
	var gate sync.Mutex
	stopping := false
	handler := server.Handler
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gate.Lock()
		if stopping {
			gate.Unlock()
			http.Error(w, "服务正在停止", http.StatusServiceUnavailable)
			return
		}
		requests.Add(1)
		gate.Unlock()
		defer requests.Done()
		handler.ServeHTTP(w, r)
	})
	server.BaseContext = func(net.Listener) context.Context { return ctx }
	backgroundDone := make(chan struct{})
	go func() { defer close(backgroundDone); background(ctx) }()
	shutdownDone := make(chan error, 1)
	go func() {
		<-ctx.Done()
		gate.Lock()
		stopping = true
		gate.Unlock()
		shutdownCtx, done := context.WithTimeout(context.Background(), grace)
		defer done()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			_ = server.Close()
		}
		requests.Wait()
		shutdownDone <- err
	}()
	err := server.Serve(listener)
	// Serve can also stop on an accept error without a cancellation signal.
	cancel()
	shutdownErr := <-shutdownDone
	<-backgroundDone
	if err == http.ErrServerClosed {
		return shutdownErr
	}
	return err
}
