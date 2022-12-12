package main

import (
	"fmt"
	"log"
	"os"

	"bytes"

	"regexp"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

func main() {
	hosts := []string{
		"10.127.0.10",
		"10.127.0.101",
		"10.127.0.102",
		"10.127.0.11",
		"10.127.0.112",
		"10.127.0.12",
		"10.127.0.133",
		"10.127.0.14",
		"10.127.0.15",
		"10.127.0.153",
		"10.127.0.154",
		"10.127.0.155",
		"10.127.0.158",
		"10.127.0.159",
		"10.127.0.200",
		"10.127.0.212",
		"10.127.0.54",
		"10.127.0.53",
		"10.127.0.55",
		"10.127.0.56",
		"10.127.0.57",
		"10.127.0.58",
		"10.127.0.59",
		"10.127.0.60",
		"10.127.0.65",
		"10.127.0.103",
		"10.127.0.104",
		// switches
		"10.128.0.37",
		"10.128.0.36",
		"10.128.0.35",
	}
	username := os.Getenv("USERNAME")
	password := os.Getenv("PW")
	if len(password) == 0 || len(username) == 0 {
		log.Print("empty username or password, try load .env")

		err := godotenv.Load(".env")

		if err != nil {
			log.Fatal("Error loading .env file", err)
		}
	}

	username = os.Getenv("USERNAME")
	password = os.Getenv("PW")
	if len(password) == 0 || len(username) == 0 {
		log.Fatal("empty username or password")
	}

	g := new(errgroup.Group)
	for _, hostname := range hosts {
		hostname := hostname // https://go.dev/doc/faq#closures_and_goroutines
		g.Go(func() error {
			return run(hostname, username, password)
		})
	}

	log.Print("doing")
	err := g.Wait()
	if err != nil {
		log.Print(err)
	} else {
		log.Print("all done no errors")
	}

}

func run(hostname, username, password string) error {
	var reConf = regexp.MustCompile(`(?s)Current configuration .*end`)
	var reHost = regexp.MustCompile(`(?m)^hostname\s([-0-9A-Za-z_]+).?$`)
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
		"aes192-cbc", "aes256-cbc", "aes128-ctr", "aes192-ctr", "aes256-ctr")

	//////////////////////////////
	// Connect to host
	client, err := ssh.Dial("tcp", hostname+":"+port, config)
	if err != nil {
		log.Print(err)
		return err
	}
	defer client.Close()

	// Create sesssion
	sess, err := client.NewSession()
	if err != nil {
		log.Print(err)
		return err
	}
	defer sess.Close()

	// StdinPipe for commands
	stdin, err := sess.StdinPipe()
	if err != nil {
		log.Print(err)
		return err
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
		log.Print(err)
		return err
	}

	// send the commands
	commands := []string{
		"terminal length 0",
		"show running-config",
		"exit",
	}
	for _, cmd := range commands {
		_, err = fmt.Fprintf(stdin, "%s\n", cmd)
		if err != nil {
			log.Print(err)
			return err
		}
	}

	// log.Print(hostname)
	// Wait for sess to finish
	err = sess.Wait()
	if err != nil {
		log.Print(err)
		return err
	}

	// Uncomment to store in variable
	// fmt.Println(b.String())
	out := b.Bytes()
	// fmt.Println(string(out))

	if reHost.Match(out) {
		fname := string(reHost.FindSubmatch(out)[1])

		log.Print(fname)

		if reConf.Match(out) {
			config := reConf.FindAll(out, -1)[0]
			err := os.WriteFile(fname, config, 0644)
			if err != nil {
				log.Print(err)
				return err
			}
		} else {
			log.Print(hostname, "config not found")
		}
	} else {
		log.Print(hostname, "hostname not found")
	}
	// time.Sleep(5 * time.Second)
	return nil
}
