package cmd

import (
	"ezlabctl/internal"
	"log"

	"github.com/spf13/cobra"
)

var attachStorageCmd = &cobra.Command{
    Use:   "attach",
    Short: "Configure external Data Fabric for UA deployment",
    Run: func(cmd *cobra.Command, args []string) {
        // Ensure required parameters
        dfhost = internal.GetStringInput(cmd, "dfhost", "EDF host", "")
        dfuser = internal.GetStringInput(cmd, "dfuser", "EDF admin user", "mapr")
        dfpass = internal.GetStringInput(cmd, "dfpass", "EDF admin password", "mapr")
        // sshuser = internal.GetStringInput(cmd, "sshuser", "SSH User", "root")
        // sshpass = internal.GetStringInput(cmd, "password", "SSH Password", "")
        PrepareEDF(cmd)

    },
}

func init() {
    unifiedAnalyticsCmd.AddCommand(attachStorageCmd)
    attachStorageCmd.Flags().StringVarP(&dfhost, "dfhost", "", "", "EDF Host")
    attachStorageCmd.Flags().StringVarP(&dfuser, "dfuser", "", "mapr", "EDF Admin User")
    attachStorageCmd.Flags().StringVarP(&dfpass, "dfpass", "", "mapr", "EDF Admin Password")
    attachStorageCmd.MarkFlagRequired("dfhost")
    attachStorageCmd.MarkFlagRequired("dfuser")
    attachStorageCmd.MarkFlagRequired("dfpass")
}

func PrepareEDF(cmd *cobra.Command) {
    log.Println("Configure Data Fabric for UA...")

    if dfhost == "" || dfuser == "" || dfpass == "" {
        log.Fatalf("Missing input\nHost: %s\nUser: %s\nPass: %s\n", dfhost, dfuser, dfpass)
    }

    filesToTransfer := []string{
        "ezua-maprtenantticket",
        "ezua-cldb-nodes.json",
        "ezua-rest-nodes.json",
        "ezua-s3-nodes.json",
        "ezua-s3-keys.json",
        "ezua-chain-ca.pem",
    }

    commands := internal.DfSetupForUACommands(dfuser, dfpass, filesToTransfer)
    // TODO: add S3 IAM policy
    // internal.SCPPutFile(dfhost, appConfig.Username, appConfig.Password, "./templates/s3_iam_policy.json", "/tmp/s3_iam_policy.json")

    log.Printf("Connect to %s...\n", dfhost)
    internal.SshCommands(dfhost, sshuser, sshpass, commands)
    log.Println("DF is configured for UA...")

    // Perform SCP transfer
    log.Println("Getting DF files to /tmp...")
    internal.ScpGetFiles(dfhost, dfuser, dfpass, "/tmp", "/tmp", filesToTransfer)

    log.Println("DF Setup is complete. Ensure DF is licensed to use NFS service")
}