package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

func makeSigner(keyname string) (signer ssh.Signer, err error) {
	fp, err := os.Open(keyname)
	if err != nil {
		return
	}
	defer fp.Close()

	buf, _ := ioutil.ReadAll(fp)
	signer, _ = ssh.ParsePrivateKey(buf)
	return
}

func getSigners() ([]ssh.Signer, error) {
	signers := []ssh.Signer{}
	keys := []string{os.Getenv("SSH_KEY_PATH")}

	for _, keyname := range keys {
		signer, err := makeSigner(keyname)
		if err == nil {
			signers = append(signers, signer)
		}
	}
	return signers, nil
}

func executeCmd(cmd, hostname string, config *ssh.ClientConfig) string {
	conn, _ := ssh.Dial("tcp", hostname+":22", config)
	session, _ := conn.NewSession()
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)

	return stdoutBuf.String()
}

func SSH(user string, hostname string, cmd string) string {

	results := make(chan string, 10)
	timeout := time.After(5 * time.Second)
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeysCallback(getSigners)},
	}

	go func(hostname string) {
		results <- executeCmd(cmd, hostname, config)
	}(hostname)

	select {
	case res := <-results:
		return res
	case <-timeout:
		return "Timed out!"
	}
}

func SSHAction(args []string) {
	s := SSH(args[0], args[1], "docker run --rm hello-world")
	fmt.Print(s)
}
