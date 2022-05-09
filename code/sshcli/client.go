package sshcli

import (
	"context"
	"io"
	"net"
	"os"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	cli *ssh.Client
}

func New(addr, user, pass string) (*Client, error) {
	cli, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(pass)},
		HostKeyCallback: func(string, net.Addr, ssh.PublicKey) error {
			return nil
		},
	})
	if err != nil {
		return nil, err
	}
	return &Client{cli}, nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) Do(ctx context.Context, cmd, stdin string) (string, error) {
	session, err := c.cli.NewSession()
	if err != nil {
		return "", err
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	err = session.RequestPty("xterm", 24, 80, modes)
	if err != nil {
		return "", err
	}
	var buf writer
	if len(stdin) > 0 {
		session.Stdin = strings.NewReader(stdin)
	}
	session.Stdout = &buf
	session.Stderr = &buf
	err = session.Start(cmd)
	if err != nil {
		return "", err
	}
	ch := make(chan error)
	go func() {
		ch <- session.Wait()
		session.Close()
	}()
	select {
	case err = <-ch:
	case <-ctx.Done():
		err = ctx.Err()
	}
	if len(stdin) > 0 {
		return strings.TrimPrefix(buf.String(), strings.ReplaceAll(stdin, "\n", "\r\n")), err
	}
	return buf.String(), err
}

func (c *Client) Upload(src, dst string) error {
	cli, err := sftp.NewClient(c.cli)
	if err != nil {
		return err
	}
	defer cli.Close()
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := cli.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UploadFrom(src io.Reader, dst string) error {
	cli, err := sftp.NewClient(c.cli)
	if err != nil {
		return err
	}
	defer cli.Close()
	dstFile, err := cli.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, src)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Remove(dir string) error {
	cli, err := sftp.NewClient(c.cli)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.Remove(dir)
}

func (c Client) Glob(pattern string) ([]string, error) {
	cli, err := sftp.NewClient(c.cli)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	return cli.Glob(pattern)
}

func (c Client) Stat(dir string) (os.FileInfo, error) {
	cli, err := sftp.NewClient(c.cli)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	return cli.Stat(dir)
}

func (c Client) StatVFS(dir string) (*sftp.StatVFS, error) {
	cli, err := sftp.NewClient(c.cli)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	return cli.StatVFS(dir)
}

func (c Client) ReadFile(dir string) ([]byte, error) {
	cli, err := sftp.NewClient(c.cli)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	f, err := cli.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
