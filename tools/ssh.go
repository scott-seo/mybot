package tools

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"

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

	fileList := []string{}

	home, _ := homedir.Dir()
	err := filepath.Walk(home+"/.aws/", func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && strings.HasSuffix(f.Name(), "pem") {
			fileList = append(fileList, path)
		}
		return nil
	})
	err = filepath.Walk(home+"/.ssh/", func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && (strings.HasSuffix(f.Name(), "_rsa") || strings.HasSuffix(f.Name(), "_dsa")) {
			fileList = append(fileList, path)
		}
		return nil
	})

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println(fileList)

	keys := fileList

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

func SSHAction(arg string) {
	args := strings.Split(arg, " ")
	user := strings.Split(args[0], "@")[0]
	host := strings.Split(args[0], "@")[1]
	cmd := strings.Replace(strings.Join(args[1:], " "), `"`, "", -1)
	s := SSH(user, host, cmd)
	fmt.Println(s)
}
