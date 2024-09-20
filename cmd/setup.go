package cmd

import (
	"ezlabctl/internal"
	"log"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var setupDfCmd = &cobra.Command{
    Use:   "setup",
    Short: "Configure external Data Fabric for UA deployment",
    Run: func(cmd *cobra.Command, args []string) {
        log.Println("Configure Data Fabric for UA...")

    // Get user input for DF
    appConfig, err := internal.GetDFInput()
    if err != nil {
        log.Fatalf("failed to get configuration for Data Fabric: %v", err)
    }
    // log.Printf("Got configuration with %s", df)

    filesToTransfer := []string{
        "maprtenantticket",
        "cldb_nodes.json",
        "rest_nodes.json",
        "s3_nodes.json",
        "s3_keys.json",
    }

    commands := []string{
        // "( id ezua && id -nG ezua | grep -s $(id -un ezua) ) || sudo useradd -m -s /bin/bash -U ezua",
        // "id ezua || sudo useradd -m -U ezua",
        "id ezua 2>&1 >/dev/null || sudo useradd -m -U ezua 2>&1 > /dev/null",
        "echo ezua: " + appConfig.DFPass + " | sudo chpasswd",
        "echo " + appConfig.DFPass + " | maprlogin password -user " + appConfig.DFUser,
        "echo Setting up the volumes for UA...",
        "maprcli acl edit -type cluster -user ezua:login,cv",
        "maprcli volume create -name ezua-base-volume -path /ezua -type rw -json -rootdiruser ezua -rootdirgroup ezua -createparent 1 || true",
        "[ -f /tmp/maprtenantticket ] || maprlogin generateticket -type tenant -user ezua -out /tmp/maprtenantticket",
        "[ -f /tmp/cldb_nodes.json ] || /opt/mapr/bin/maprcli node list -columns hn -filter svc==cldb -json > /tmp/cldb_nodes.json",
        "[ -f /tmp/rest_nodes.json ] || /opt/mapr/bin/maprcli node list -columns hn -filter svc==apiserver -json > /tmp/rest_nodes.json",
        "[ -f /tmp/s3_nodes.json ] || /opt/mapr/bin/maprcli node list -columns hn -filter svc==s3server -json > /tmp/s3_nodes.json",
        "[ -f /tmp/s3_keys.json ] || maprcli s3keys generate -domainname primary -accountname default -username ezua -json > /tmp/s3_keys.json",
        "sudo chown mapr:mapr /tmp/" + strings.Join(filesToTransfer, " /tmp/"),
    }

    var wg sync.WaitGroup // Create a WaitGroup
    wg.Add(1)
    go internal.SSHCommands(appConfig.DFHost, appConfig.Username, appConfig.Password, commands, &wg)
    wg.Wait()

    // Perform SCP transfer
	go internal.SCPGetFiles(appConfig.DFHost, appConfig.DFUser, appConfig.DFPass, "/tmp", "/tmp", filesToTransfer)
    log.Println("Files transferred.")

    log.Println("DF Setup is done!")
    },
}
