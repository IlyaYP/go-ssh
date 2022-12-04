package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func main() {

	user := "suprunis"
	NumberOfPrompts := 3

	// Normally this would be a callback that prompts the user to answer the
	// provided questions
	Cb := func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
		pw, err := credentials()
		return []string{pw}, err
	}

	config := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            user,
		Auth: []ssh.AuthMethod{
			// ssh.RetryableAuthMethod(ssh.KeyboardInteractiveChallenge(Cb), NumberOfPrompts),
			// ssh.Password(pass), // works
			// ssh.PasswordCallback(credentials), // works
			ssh.RetryableAuthMethod(ssh.PasswordCallback(credentials), NumberOfPrompts), // works
		},
		// HostKeyAlgorithms: []string{},
	}

	// host := "10.128.0.36:22"
	host := "192.168.66.156:22"

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer client.Close()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("dir"); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())

	// netConn, err := net.Dial("tcp", host)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// sshConn, _, _, err := ssh.NewClientConn(netConn, host, config)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _ = sshConn
	_, _ = Cb, NumberOfPrompts
}

func credentials() (string, error) {
	// reader := bufio.NewReader(os.Stdin)

	// fmt.Print("Enter Username: ")
	// username, err := reader.ReadString('\n')
	// if err != nil {
	// 	return "", "", err
	// }

	fmt.Print("\nEnter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	password := string(bytePassword)
	return strings.TrimSpace(password), nil
}
