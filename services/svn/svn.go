// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const MaxOutput = 16 << 20

type CommandError struct{ Tool, Reason string }

func (e *CommandError) Error() string { return fmt.Sprintf("%s failed (%s)", e.Tool, e.Reason) }
func (e *CommandError) UserMessage() string {
	for _, cause := range []struct{ code, message string }{
		{"E215004", "远端SVN认证失败，请核对远端用户名、密码及仓库读取权限"},
		{"E170001", "远端SVN授权失败，请核对仓库读取权限"},
		{"E230001", "远端HTTPS证书校验失败，请检查证书配置"},
		{"E175013", "远端SVN服务器拒绝访问"},
		{"E670002", "无法解析远端SVN主机名"},
		{"E730061", "远端SVN服务器拒绝连接"},
		{"E170013", "无法连接远端SVN，请检查认证、证书和网络"},
		{"E175002", "远端SVN请求失败，请检查服务器状态和网络连接"},
		{"E160006", "请求的修订不存在，请检查来源仓库是否发生修订回退"},
		{"E155007", "指定路径不是有效的 SVN 工作副本，请检查操作路径"},
		{"E200015", "SVN操作被中断，请检查任务状态后重试"},
		{"E180001", "无法打开 SVN 仓库，请检查路径、仓库格式及服务账号权限"},
	} {
		if strings.Contains(e.Reason, cause.code) {
			return cause.message + "（" + cause.code + "）"
		}
	}
	return "SVN操作失败，当前记录没有可识别的错误代码；请检查仓库设置、服务账号权限及运行日志"
}

type Runner struct {
	OnOutput  func()
	OnCommand func(string)
	Bin       string
	Timeout   time.Duration
}
type bounded struct {
	bytes.Buffer
	limit    int
	exceeded bool
}

func (b *bounded) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		b.exceeded = true
		return 0, errors.New("command output exceeds limit")
	}
	return b.Buffer.Write(p)
}
func (r Runner) Run(ctx context.Context, tool string, input io.Reader, args ...string) ([]byte, error) {
	return r.run(ctx, tool, input, nil, args...)
}

func (r Runner) run(ctx context.Context, tool string, input io.Reader, onLine func(string) error, args ...string) ([]byte, error) {
	switch tool {
	case "svn", "svnadmin", "svnlook", "svnsync", "svnrdump":
	default:
		return nil, errors.New("unsupported tool")
	}
	if r.OnCommand != nil {
		r.OnCommand(tool)
	}
	timeout := r.Timeout
	if timeout == 0 {
		timeout = 30 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	executable := filepath.Join(r.Bin, tool)
	if filepath.Ext(executable) == "" {
		executable = platformExecutable(executable)
	}
	if tool == "svn" || tool == "svnsync" || tool == "svnrdump" {
		config := filepath.Join(filepath.Dir(r.Bin), "svn-config")
		if err := os.MkdirAll(config, 0700); err != nil {
			return nil, err
		}
		args = append(args, "--config-dir", config)
	}
	cmd := exec.CommandContext(ctx, executable, args...)
	hideWindow(cmd)
	cmd.Stdin = input
	out := &bounded{limit: MaxOutput}
	stderr := &bounded{limit: 64 << 10}
	var observer *lineOutput
	var stdout io.Writer = out
	if onLine != nil {
		cmd.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
		observer = &lineOutput{line: onLine, cancel: cancel}
		stdout = observer
	}
	cmd.Stdout = progressWriter{Writer: stdout, notify: r.OnOutput}
	cmd.Stderr = progressWriter{Writer: stderr, notify: r.OnOutput}
	err := cmd.Run()
	if observer != nil {
		observer.finish()
		if observer.err != nil {
			return nil, observer.err
		}
	}
	if err != nil {
		if out.exceeded || stderr.exceeded {
			return nil, errors.New("command output exceeds limit")
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, &CommandError{Tool: tool, Reason: safeError(stderr.String())}
	}
	return out.Bytes(), nil
}
func safeError(s string) string {
	// SVN diagnostics may contain remote URLs and authentication information.
	// Prefer the actionable inner cause over a generic E170013 wrapper.
	for _, cause := range []struct{ match, summary string }{
		{"E215004", "E215004: authentication failed; verify the remote account and repository access"},
		{"E170001", "E170001: authorization failed; verify repository read access"},
		{"E230001", "E230001: TLS certificate verification failed"},
		{"E175013", "E175013: remote server denied access"},
		{"E670002", "E670002: remote host name could not be resolved"},
		{"E730061", "E730061: remote server refused the connection"},
	} {
		if strings.Contains(s, cause.match) {
			return cause.summary
		}
	}
	for _, code := range []string{"E170001", "E170013", "E175002", "E160006", "E155007", "E200015", "E180001"} {
		if strings.Contains(s, code) {
			return code
		}
	}
	return "see server diagnostics; no credentials recorded"
}

type Credentials struct {
	Username, Password string
	SkipTLS            bool
}

func (c Credentials) Args() []string {
	a := []string{"--non-interactive", "--no-auth-cache"}
	if c.Username != "" {
		a = append(a, "--username", c.Username)
	}
	if c.Password != "" {
		a = append(a, "--password-from-stdin")
	}
	if c.SkipTLS {
		a = append(a, "--trust-server-cert-failures=unknown-ca,cn-mismatch,expired,not-yet-valid,other")
	}
	return a
}
func (c Credentials) Input() io.Reader {
	if c.Password == "" {
		return nil
	}
	// SVN consumes the supplied password bytes; an added newline changes it.
	return strings.NewReader(c.Password)
}

// svnsync does not implement svn's --password-from-stdin option. Never log its argv.
func (c Credentials) SyncArgs() []string { return c.syncArgs(true) }
func (c Credentials) syncArgs(quiet bool) []string {
	a := []string{"--non-interactive", "--no-auth-cache", "--steal-lock"}
	if quiet {
		a = append(a, "--quiet")
	}
	if c.Username != "" {
		a = append(a, "--source-username", c.Username)
	}
	if c.Password != "" {
		a = append(a, "--source-password", c.Password)
	}
	if c.SkipTLS {
		a = append(a, "--source-trust-server-cert-failures=unknown-ca,cn-mismatch,expired,not-yet-valid,other")
	}
	return a
}
func ValidateRemote(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("expected an HTTP/HTTPS repository URL without credentials, query or fragment")
	}
	return nil
}

type Info struct {
	Entries []struct {
		Revision   int64  `xml:"revision,attr"`
		URL        string `xml:"url"`
		Repository struct {
			Root string `xml:"root"`
			UUID string `xml:"uuid"`
		} `xml:"repository"`
	} `xml:"entry"`
}

func (r Runner) Info(ctx context.Context, target string, c Credentials) (string, string, int64, error) {
	b, err := r.Run(ctx, "svn", c.Input(), append([]string{"info", "--xml", "-r", "HEAD", target + "@"}, c.Args()...)...)
	if err != nil {
		return "", "", 0, err
	}
	var info Info
	if err = xml.Unmarshal(b, &info); err != nil || len(info.Entries) != 1 {
		return "", "", 0, errors.New("invalid SVN info response")
	}
	e := info.Entries[0]
	return e.Repository.UUID, e.Repository.Root, e.Revision, nil
}
func LocalURL(p string) string {
	p = filepath.ToSlash(p)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return (&url.URL{Scheme: "file", Path: p}).String()
}
func Target(root, relative string) (string, error) {
	if strings.ContainsRune(relative, 0) || strings.Contains(relative, "\\") {
		return "", errors.New("invalid repository path")
	}
	for _, part := range strings.Split(relative, "/") {
		if part == ".." {
			return "", errors.New("invalid repository path")
		}
	}
	u, err := url.Parse(LocalURL(root))
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, relative)
	return u.String() + "@", nil
}
func (r Runner) LocalInfo(ctx context.Context, p string) (string, int64, error) {
	b, err := r.Run(ctx, "svnlook", nil, "uuid", p)
	if err != nil {
		return "", 0, err
	}
	uuid := strings.TrimSpace(string(b))
	b, err = r.Run(ctx, "svnlook", nil, "youngest", p)
	if err != nil {
		return "", 0, err
	}
	rev, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	return uuid, rev, err
}
func (r Runner) CreateMirror(ctx context.Context, dest, remote string, c Credentials) error {
	if err := ValidateRemote(remote); err != nil {
		return err
	}
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		if _, err = r.Run(ctx, "svnadmin", nil, "create", dest); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if b, err := r.Run(ctx, "svnlook", nil, "propget", "--revprop", "-r", "0", dest, "svn:sync-from-url"); err == nil {
		if strings.TrimSpace(string(b)) != remote {
			return errors.New("existing mirror source mismatch")
		}
		return nil
	}
	_, rev, err := r.LocalInfo(ctx, dest)
	if err != nil {
		return err
	}
	if rev != 0 {
		return errors.New("uninitialized mirror contains revisions")
	}
	hook := filepath.Join(dest, "hooks", hookFilename())
	if err := os.WriteFile(hook, []byte(hookContent()), 0700); err != nil {
		return err
	}
	uuid, _, _, err := r.Info(ctx, remote, c)
	if err != nil {
		return err
	}
	if _, err = r.Run(ctx, "svnadmin", nil, "setuuid", dest, uuid); err != nil {
		return err
	}
	a := append([]string{"initialize", LocalURL(dest), remote}, c.SyncArgs()...)
	_, err = r.Run(ctx, "svnsync", nil, a...)
	return err
}
func (r Runner) Sync(ctx context.Context, dest string, c Credentials) error {
	_, err := r.Run(ctx, "svnsync", nil, append([]string{"synchronize", LocalURL(dest)}, c.SyncArgs()...)...)
	return err
}
func (r Runner) SetRemote(ctx context.Context, dest, remote string) error {
	_, err := r.Run(ctx, "svn", nil, "propset", "--revprop", "-r", "0", "svn:sync-from-url", remote, LocalURL(dest), "--non-interactive")
	return err
}

type Entry struct {
	Kind   string `xml:"kind,attr"`
	Name   string `xml:"name"`
	Size   int64  `xml:"size"`
	Commit struct {
		Revision int64  `xml:"revision,attr"`
		Author   string `xml:"author"`
		Date     string `xml:"date"`
	} `xml:"commit"`
}

func (r Runner) List(ctx context.Context, root, p string, rev int64) ([]Entry, error) {
	t, err := Target(root, p)
	t += strconv.FormatInt(rev, 10)
	if err != nil {
		return nil, err
	}
	b, err := r.Run(ctx, "svn", nil, "list", "--xml", "-r", strconv.FormatInt(rev, 10), t, "--non-interactive")
	if err != nil {
		return nil, err
	}
	var result struct {
		List struct {
			Entries []Entry `xml:"entry"`
		} `xml:"list"`
	}
	err = xml.Unmarshal(b, &result)
	return result.List.Entries, err
}

type LogEntry struct {
	Revision int64  `xml:"revision,attr"`
	Author   string `xml:"author"`
	Date     string `xml:"date"`
	Message  string `xml:"msg"`
	Paths    []struct {
		CopyFromPath     string `xml:"copyfrom-path,attr"`
		CopyFromRevision int64  `xml:"copyfrom-rev,attr"`
		Action           string `xml:"action,attr"`
		Path             string `xml:",chardata"`
	} `xml:"paths>path"`
}

func (r Runner) Log(ctx context.Context, root, p string, rev int64) ([]LogEntry, error) {
	t, err := Target(root, p)
	t += strconv.FormatInt(rev, 10)
	if err != nil {
		return nil, err
	}
	b, err := r.Run(ctx, "svn", nil, "log", "--xml", "-v", "--limit", "100", "-r", fmt.Sprintf("%d:0", rev), t, "--non-interactive")
	if err != nil {
		return nil, err
	}
	var result struct {
		Entries []LogEntry `xml:"logentry"`
	}
	err = xml.Unmarshal(b, &result)
	return result.Entries, err
}

// HistoryPage fixes the peg revision so later renames do not change the requested history.
func (r Runner) HistoryPage(ctx context.Context, root, p string, head, cursor int64, newer bool, limit int) ([]LogEntry, error) {
	if head == 0 && cursor == 0 {
		return []LogEntry{}, nil
	}
	if head < 1 || cursor < 1 || cursor > head || limit < 1 || limit > 31 {
		return nil, errors.New("invalid history cursor")
	}
	target, err := Target(root, p)
	if err != nil {
		return nil, err
	}
	target += strconv.FormatInt(head, 10)
	end := int64(1)
	if newer {
		end = head
	}
	b, err := r.Run(ctx, "svn", nil, "log", "--xml", "-v", "--limit", strconv.Itoa(limit), "-r", fmt.Sprintf("%d:%d", cursor, end), target, "--non-interactive")
	if err != nil {
		return nil, err
	}
	var result struct {
		Entries []LogEntry `xml:"logentry"`
	}
	err = xml.Unmarshal(b, &result)
	return result.Entries, err
}

func (r Runner) Cat(ctx context.Context, root, p string, rev int64) ([]byte, error) {
	t, err := Target(root, p)
	t += strconv.FormatInt(rev, 10)
	if err != nil {
		return nil, err
	}
	return r.Run(ctx, "svn", nil, "cat", "-r", strconv.FormatInt(rev, 10), t, "--non-interactive")
}

// LogPage reads a bounded page at the published revision, without expensive changed-path lists.
func (r Runner) LogPage(ctx context.Context, root string, head, next int64, limit int) ([]LogEntry, error) {
	if head < 1 || next < 1 || next > head || limit < 1 || limit > 500 {
		return nil, errors.New("invalid log page")
	}
	target, err := Target(root, "")
	if err != nil {
		return nil, err
	}
	target += strconv.FormatInt(head, 10)
	b, err := r.Run(ctx, "svn", nil, "log", "--xml", "--limit", strconv.Itoa(limit), "-r", fmt.Sprintf("%d:1", next), target, "--non-interactive")
	if err != nil {
		return nil, err
	}
	var result struct {
		Entries []LogEntry `xml:"logentry"`
	}
	err = xml.Unmarshal(b, &result)
	return result.Entries, err
}

type progressWriter struct {
	io.Writer
	notify func()
}

func (w progressWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	if n > 0 && w.notify != nil {
		w.notify()
	}
	return n, err
}
