package cmd

import (
	"ezlabctl/internal"
	"log"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var storageCmd = &cobra.Command{
    Use:   "storage",
    Short: "Configure external Data Fabric for UA deployment",
    Run: func(cmd *cobra.Command, args []string) {

        log.Println("Configure Data Fabric for UA...")

    // Get user input for DF
    // appConfig, err := internal.GetDFInput()
    // if err != nil {
    //     log.Fatalf("failed to get configuration for Data Fabric: %v", err)
    // }
    appConfig := internal.GetAppConfiguration()

    var dfhost, dfuser, dfpass string
    var err error

    if cmd.Flags().Changed("server") {
        dfhost, err = cmd.Flags().GetString("server")
        if err != nil {
            log.Fatal("failed to get server value: %w", err)
        }
    } else {
        dfhost = internal.AskForInput("DF Server: ", "")
    }

    if cmd.Flags().Changed("username") {
        dfuser, err = cmd.Flags().GetString("username")
        if err != nil {
            log.Fatal("failed to get username value: %w", err)
        }
    } else {
        dfuser = internal.AskForInput("Username: ", "mapr")
    }

    if cmd.Flags().Changed("password") {
        dfpass, err = cmd.Flags().GetString("password")
        if err != nil {
            log.Fatal("failed to get password value: %w", err)
        }
    } else {
        dfpass = internal.AskForInput("Password: ", "mapr")
    }

    if dfhost == "" || dfuser == "" || dfpass == "" {
        log.Fatalf("Missing input\nHost: %s\nUser: %s\nPass: %s\n", dfhost, dfuser, dfpass)
    }

    filesToTransfer := []string{
        "maprtenantticket",
        "cldb_nodes.json",
        "rest_nodes.json",
        "s3_nodes.json",
        "s3_keys.json",
    }

    commands := []string{
        "id ezua || sudo useradd -m -U ezua",
        "echo ezua:" + dfpass + " | sudo chpasswd",
        "[ -f /tmp/maprticket_$(id -u) ] || echo " + dfpass + " | maprlogin password -user " + dfuser,
        "echo Setting up the UA volumes...",
        "maprcli acl edit -type cluster -user ezua:login,cv",
        "maprcli volume create -name ezua-base-volume -path /ezua -type rw -json -rootdiruser ezua -rootdirgroup ezua -createparent 1 || true",
        "[ -f /tmp/maprtenantticket ] || maprlogin generateticket -type tenant -user ezua -out /tmp/maprtenantticket",
        "[ -f /tmp/cldb_nodes.json ] || /opt/mapr/bin/maprcli node list -columns hn -filter svc==cldb -json > /tmp/cldb_nodes.json",
        "[ -f /tmp/rest_nodes.json ] || /opt/mapr/bin/maprcli node list -columns hn -filter svc==apiserver -json > /tmp/rest_nodes.json",
        "[ -f /tmp/s3_nodes.json ] || /opt/mapr/bin/maprcli node list -columns hn -filter svc==s3server -json > /tmp/s3_nodes.json",
        "[ -f /tmp/s3_keys.json ] || maprcli s3keys generate -domainname primary -accountname default -username ezua -json > /tmp/s3_keys.json",
        "sudo chown mapr:mapr /tmp/" + strings.Join(filesToTransfer, " /tmp/"),
        "[ -d /mapr ] || sudo mkdir /mapr",
        "mount | grep -q /mapr || sudo mount -t nfs -o vers=3,nolock " + dfhost + ":/mapr/ /mapr",
    }

    // TODO: add S3 IAM policy
    // internal.SCPPutFile(dfhost, appConfig.Username, appConfig.Password, "./templates/s3_iam_policy.json", "/tmp/s3_iam_policy.json")

    var wg sync.WaitGroup // Create a WaitGroup
    wg.Add(1)
    log.Println("Setting up the DF for UA...")
    go internal.SSHCommands(dfhost, appConfig.Username, appConfig.Password, commands, &wg)
    wg.Wait()
    log.Println("DF is configured for UA...")

    // Perform SCP transfer
    log.Println("Getting DF files to /tmp...")
	internal.SCPGetFiles(dfhost, dfuser, dfpass, "/tmp", "/tmp", filesToTransfer)

    log.Println("DF Setup is complete. Add license to enable NFS service!")
    },
}

func init() {
    storageCmd.Flags().StringP("server", "s", "", "Hostname/IP")
    storageCmd.Flags().StringP("username", "u", "mapr", "Username")
    storageCmd.Flags().StringP("password", "p", "mapr", "Password")
    rootCmd.AddCommand(storageCmd)
}