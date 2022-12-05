package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func main() {

	user := "suprunis"
	NumberOfPrompts := 3

	config := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            user,
		Auth: []ssh.AuthMethod{
			ssh.RetryableAuthMethod(ssh.PasswordCallback(credentials), NumberOfPrompts), // works
		},
	}

	config.KeyExchanges = append(
		config.KeyExchanges,
		"diffie-hellman-group-exchange-sha256",
		"diffie-hellman-group-exchange-sha1",
		"diffie-hellman-group1-sha1",
	)

	config.Ciphers = append(config.Ciphers, "aes128-cbc", "3des-cbc",
		"aes192-cbc", "aes256-cbc")

	host := "10.128.0.36:22"
	// host := "192.168.66.156:22"

	/////////////////////////////////////

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

	var stdin io.WriteCloser
	// var stdin io.Writer
	var stdout, stderr io.Reader

	stdin, err = session.StdinPipe()
	if err != nil {
		fmt.Println(err.Error())
	}

	stdout, err = session.StdoutPipe()
	if err != nil {
		fmt.Println(err.Error())
	}

	stderr, err = session.StderrPipe()
	if err != nil {
		fmt.Println(err.Error())
	}

	wr := make(chan []byte, 10)

	go func() {
		for d := range wr {
			_, err := stdin.Write(d)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for {
			if tkn := scanner.Scan(); tkn {
				rcv := scanner.Bytes()

				raw := make([]byte, len(rcv))
				copy(raw, rcv)

				fmt.Println(string(raw))
			} else {
				if scanner.Err() != nil {
					fmt.Println(scanner.Err())
				} else {
					fmt.Println("io.EOF")
				}
				return
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	session.Shell()

	for {
		fmt.Println("$")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		text := scanner.Text()

		wr <- []byte(text + "\n")
	}

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
