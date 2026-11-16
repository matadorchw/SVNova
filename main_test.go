// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

func TestEmbeddedFrontendAssetsMatchBuild(t *testing.T) {
	assets, err := fs.Sub(resources, "public/assets")
	if err != nil {
		t.Fatal(err)
	}
	err = fs.WalkDir(os.DirFS("public/assets"), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		built, err := os.ReadFile("public/assets/" + path)
		if err != nil {
			return err
		}
		embedded, err := fs.ReadFile(assets, path)
		if err != nil {
			t.Errorf("built asset missing from EXE: %s: %v", path, err)
			return nil
		}
		if !bytes.Equal(built, embedded) {
			t.Errorf("embedded asset differs from build: %s", path)
		}
		response := httptest.NewRecorder()
		http.FileServer(http.FS(assets)).ServeHTTP(response, httptest.NewRequest("GET", "/"+path, nil))
		if response.Code != http.StatusOK {
			t.Errorf("asset %s returned HTTP %d", path, response.Code)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

type pipeListener struct {
	connections chan net.Conn
	closed      chan struct{}
	once        sync.Once
}

func (l *pipeListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.connections:
		return c, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}
func (l *pipeListener) Close() error   { l.once.Do(func() { close(l.closed) }); return nil }
func (l *pipeListener) Addr() net.Addr { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 80} }
func testHTTP(t *testing.T) (*pipeListener, *http.Client) {
	t.Helper()
	l := &pipeListener{connections: make(chan net.Conn), closed: make(chan struct{})}
	transport := &http.Transport{DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		client, server := net.Pipe()
		select {
		case l.connections <- server:
			return client, nil
		case <-ctx.Done():
			client.Close()
			server.Close()
			return nil, ctx.Err()
		case <-l.closed:
			client.Close()
			server.Close()
			return nil, net.ErrClosed
		}
	}}
	t.Cleanup(func() { l.Close(); transport.CloseIdleConnections() })
	return l, &http.Client{Transport: transport, Timeout: 5 * time.Second}
}

func receive[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for lifecycle event")
		var zero T
		return zero
	}
}
func assertRunning(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		t.Fatalf("server returned before cleanup completed: %v", err)
	default:
	}
}

func TestShutdownWaitsForRequestsAndBackground(t *testing.T) {
	listener, client := testHTTP(t)
	defer listener.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	requestStarted := make(chan struct{})
	requestCanceled := make(chan struct{})
	releaseRequest := make(chan struct{})
	requestFinished := make(chan struct{})
	backgroundCanceled := make(chan struct{})
	releaseBackground := make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-r.Context().Done()
		close(requestCanceled)
		<-releaseRequest
		w.WriteHeader(http.StatusOK)
		close(requestFinished)
	})}
	done := make(chan error, 1)
	go func() {
		done <- serveUntilStopped(ctx, server, listener, func(ctx context.Context) { <-ctx.Done(); close(backgroundCanceled); <-releaseBackground }, 2*time.Second)
	}()
	clientDone := make(chan struct{})
	go func() {
		defer close(clientDone)
		response, e := client.Get("http://svnova.test/")
		if e == nil {
			response.Body.Close()
		}
	}()
	receive(t, requestStarted)
	cancel()
	receive(t, requestCanceled)
	receive(t, backgroundCanceled)
	assertRunning(t, done)
	close(releaseRequest)
	receive(t, requestFinished)
	receive(t, clientDone)
	assertRunning(t, done)
	close(releaseBackground)
	if err := receive(t, done); err != nil {
		t.Fatal(err)
	}
}

func TestShutdownTimeoutStillWaitsForHandlerCleanup(t *testing.T) {
	listener, client := testHTTP(t)
	defer listener.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	release := make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-release })}
	done := make(chan error, 1)
	go func() {
		done <- serveUntilStopped(ctx, server, listener, func(ctx context.Context) { <-ctx.Done() }, time.Millisecond)
	}()
	clientDone := make(chan struct{})
	go func() {
		defer close(clientDone)
		response, e := client.Get("http://svnova.test/")
		if e == nil {
			response.Body.Close()
		}
	}()
	receive(t, started)
	cancel()
	receive(t, clientDone)
	assertRunning(t, done)
	close(release)
	if err := receive(t, done); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected shutdown deadline, got %v", err)
	}
}

func TestAcceptFailureCancelsAndWaitsForBackground(t *testing.T) {
	listener, _ := testHTTP(t)
	started := make(chan struct{})
	canceled := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- serveUntilStopped(context.Background(), &http.Server{Handler: http.NotFoundHandler()}, listener, func(ctx context.Context) { close(started); <-ctx.Done(); close(canceled); <-release }, time.Second)
	}()
	receive(t, started)
	_ = listener.Close()
	receive(t, canceled)
	assertRunning(t, done)
	close(release)
	if err := receive(t, done); err == nil {
		t.Fatal("accept failure was discarded")
	}
}
