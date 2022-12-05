package main

import (
	"fmt"
	"log"
	"os"

	"bytes"

	"regexp"

	"golang.org/x/crypto/ssh"
)

func main() {

	var reConf = regexp.MustCompile(`(?s)Current configuration .*end`)
	var reHost = regexp.MustCompile(`(?m)hostname\s([-0-9A-Za-z_]+)`)
	username := os.Getenv("USERNAME")
	password := os.Getenv("PW")
	// log.Println(username, password)
	hostname := "10.128.0.36"
	// hosts := []string{"1", "2", "3"}
	port := "22"

	// SSH client config
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		// Non-production only
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	config.KeyExchanges = append(
		config.KeyExchanges,
		"diffie-hellman-group-exchange-sha256",
		"diffie-hellman-group-exchange-sha1",
		"diffie-hellman-group1-sha1",
	)

	config.Ciphers = append(config.Ciphers, "aes128-cbc", "3des-cbc",
		"aes192-cbc", "aes256-cbc")

	// Connect to host
	client, err := ssh.Dial("tcp", hostname+":"+port, config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Create sesssion
	sess, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer sess.Close()

	// StdinPipe for commands
	stdin, err := sess.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	// Uncomment to store output in variable
	var b bytes.Buffer
	sess.Stdout = &b
	//sess.Stderr = &b

	// Enable system stdout
	// Comment these if you uncomment to store in variable
	// sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	// Start remote shell
	err = sess.Shell()
	if err != nil {
		log.Fatal(err)
	}

	// send the commands
	commands := []string{
		"sh clock",
		"sh run",
		"exit",
	}
	for _, cmd := range commands {
		_, err = fmt.Fprintf(stdin, "%s\n", cmd)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Wait for sess to finish
	err = sess.Wait()
	if err != nil {
		log.Fatal(err)
	}

	// Uncomment to store in variable
	// fmt.Println(b.String())
	out := b.Bytes()
	// fmt.Println(string(out))

	if reHost.Match(out) {
		fname := string(reHost.FindSubmatch(out)[1])

		fmt.Println(fname)

		if reConf.Match(out) {
			config := reConf.FindAll(out, -1)[0]
			fmt.Println(string(config))
			err := os.WriteFile(fname, config, 0644)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("NO")
		}
	} else {
		fmt.Println("NO")
	}

}
