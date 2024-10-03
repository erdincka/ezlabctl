package internal

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

// SSHCommand runs a single ssh command on a remote system
func SSHCommand(host string, user string, password string, cmd string) error {
    // DEBUG
    // log.Printf("Connecting to %s as user %s with password %s\n", host, user, password)
    config := &ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
            ssh.Password(password),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    client, err := ssh.Dial("tcp", host+":22", config)
    if err != nil {
        return fmt.Errorf("failed to dial: %v", err)
    }
    defer client.Close()

    session, err := client.NewSession()
    if err != nil {
        return fmt.Errorf("failed to create session: %v", err)
    }
    defer session.Close()

    // session.RequestPty()

    var stdout bytes.Buffer
    var stderr bytes.Buffer
    session.Stdout = &stdout
    session.Stderr = &stderr

    if err := session.Run(cmd); err != nil {
        log.Printf("failed command %s: %v", cmd, err)
        return err
    }

    resultErr := strings.TrimSpace(stderr.String())
    if len(resultErr) > 0 {
        log.Println("STDERR:", resultErr)
    }

    resultOut := strings.TrimSpace(stdout.String())
    if len(resultOut) > 0 {
        log.Println(host, ">>>",  resultOut)
    }

    return nil
}

// SSHCommands runs a list of ssh commands on multiple remote systems in parallel using goroutines
func SSHCommands(host string, user string, password string, cmds []string, wg *sync.WaitGroup) error {
    for _, cmd := range cmds {
        err := SSHCommand(host, user, password, cmd)
        if err != nil {
            log.Fatalln("Error running:", host, ":", err)
        }
    }
    // Notify when finished with all commmands
    wg.Done()
    return nil
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

    for _, file := range fileList {
		// Construct the remote file path
		remoteFile := filepath.Join(sourceDir, file)
		// Construct the local file path
		localFile := filepath.Join(destinationDir, file)

		// SCP command
        command := fmt.Sprintf("sshpass -f %s scp -o StrictHostKeyChecking=no %s@%s:%s %s", passfile, user, host, remoteFile, localFile)
        // log.Println("Copying file", file)
		// Execute the SCP command
        exitCode, err := RunCommand(command)
        if err != nil {
            log.Fatalf("failed to copy file:%s %v", command, err)
        } else {
            log.Printf("scp command: %s returned exit code: %d\n", command, exitCode)
        }
	}

    // Remove the password file
    err = os.Remove(passfile)
    if err != nil {
        log.Fatal(err)
    }
}

func SCPPutFile(host, user, pass, sourceFile, destinationFile string) {

    // Save password to file for later use
    passfile := ".sshpass"
    file, err := os.Create(passfile)
    if err != nil {
        log.Fatalf("%v", err)
    }
    defer file.Close()

    fmt.Fprint(file, pass)

    // SCP command
    command := fmt.Sprintf("sshpass -f %s scp -o StrictHostKeyChecking=no %s %s@%s:%s", passfile, sourceFile, user, host, destinationFile)
    log.Println("Copying file", file)
    // Execute the SCP command
    exitCode, err := RunCommand(command)
    if err != nil {
        log.Fatalf("failed to copy file:%s %v", command, err)
    } else {
        log.Printf("scp command: %s returned exit code: %d\n", command, exitCode)
    }

    // Remove the password file
    err = os.Remove(passfile)
    if err != nil {
        log.Fatal(err)
    }
}

// Get preferred outbound ip of this machine
func GetOutboundIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP.String()
}
