package internal

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

// SSHCommands connects to the remote system via SSH and runs basic commands
func SSHCommands(host, user, password string, commands []string, wg *sync.WaitGroup) {
    defer wg.Done() // Signal that the goroutine is done

    config := &ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
            ssh.Password(password),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    client, err := ssh.Dial("tcp", host+":22", config)
    if err != nil {
        log.Fatalf("failed to dial: %v", err)
    }
    defer client.Close()

    for _, cmd := range commands {
        log.Println("Running command:", cmd)
        session, err := client.NewSession()
        if err != nil {
            log.Fatalf("failed to create session: %v", err)
        }
        defer session.Close()

        // session.RequestPty()

        // output, err := session.CombinedOutput(cmd)
        var stdout bytes.Buffer
        var stderr bytes.Buffer
        session.Stdout = &stdout
        session.Stderr = &stderr

        err = session.Run(cmd)
        resultErr := strings.TrimSpace(stderr.String())
        if len(resultErr) > 0 {
            log.Println("STDERR:", resultErr)
        }
        if err != nil {
            log.Fatalf("failed command %s: %v", cmd, err)
        }

        resultOut := strings.TrimSpace(stdout.String())
        if len(resultOut) > 0 {
            log.Println(host, ">>>",  resultOut)
        }
    }
}

// Function to SCP files matching the list from a remote host
func SCPGetFiles(host, user, pass, sourceDir, destinationDir string, fileList []string) {

    // Save password to file for later use
    passfile := ".sshpass"
    file, err := os.Create(passfile)
    if err != nil {
        log.Fatalf("%v", err)
    }
    defer file.Close()

    fmt.Fprint(file, pass)

    channels := []chan string{}

    for _, file := range fileList {
        channel := make(chan string)
        channels = append(channels, channel)
		// Construct the remote file path
		remoteFile := filepath.Join(sourceDir, file)
		// Construct the local file path
		localFile := filepath.Join(destinationDir, file)

		// SCP command
        command := fmt.Sprintf("sshpass -f %s scp -o StrictHostKeyChecking=no %s@%s:%s %s", passfile, user, host, remoteFile, localFile)
        // log.Println("Copying file", file)
		// Execute the SCP command
        go ExecCommand(command, channel)
        if err != nil {
            log.Fatalf("failed to run command:%s %v", command, err)
        }
	}
    for _, channel := range channels {
        out := <-channel
        if len(out) > 0 {
            log.Println(">>>", out)
        }
    }
    // Remove the password file
    err = os.Remove(passfile)
    if err != nil {
        log.Fatal(err)
    }
}
