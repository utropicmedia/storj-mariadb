package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
)

// storeCmd represents the store command
var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Command to upload data to storj V3 network.",
	Long:  `Command to connect and transfer ALL tables from a desired MariaDB instance to given Storj Bucket.`,
	Run:   mariaStore,
}

func init() {

	// Setup the store command with its flags.
	rootCmd.AddCommand(storeCmd)
	var defaultMariaFile string
	var defaultStorjFile string
	storeCmd.Flags().BoolP("accesskey", "a", false, "Connect to storj using access key(default connection method is by using API Key).")
	storeCmd.Flags().BoolP("share", "s", false, "For generating share access of the uploaded backup file.")
	storeCmd.Flags().StringVarP(&defaultMariaFile, "maria", "i", "././config/db_property.json", "full filepath contaning MariaDB configuration.")
	storeCmd.Flags().StringVarP(&defaultStorjFile, "storj", "u", "././config/storj_config.json", "full filepath contaning storj V3 configuration.")
}

func mariaStore(cmd *cobra.Command, args []string) {

	// Process arguments from the CLI.
	mariaConfigfilePath, _ := cmd.Flags().GetString("maria")
	fullFileNameStorj, _ := cmd.Flags().GetString("storj")
	useAccessKey, _ := cmd.Flags().GetBool("accesskey")
	useAccessShare, _ := cmd.Flags().GetBool("share")

	// Read MariaDB instance's configurations from an external file and create an MariaDB configuration object.
	configMariaDB := LoadMariaProperty(mariaConfigfilePath)

	// Read storj network configurations from and external file and create a storj configuration object.
	storjConfig := LoadStorjConfiguration(fullFileNameStorj)

	// Connect to storj network using the specified credentials.
	access, project := ConnectToStorj(fullFileNameStorj, storjConfig, useAccessKey)

	// Establish connection with MariaDB and get back-up data to be uploaded.
	//	dataToUpload := CreateBackup(configMariaDB)
	reader := CreateBackup(configMariaDB)

	// Fetch all backup files from MariaDB instance and simultaneously store them into desired Storj bucket.
	fmt.Printf("\nInitiating back-up.\n")
	uploadFileName := path.Join(configMariaDB.Database, configMariaDB.Database+".sql")
	UploadData(project, storjConfig, uploadFileName, reader)
	fmt.Printf("\nBack-up complete.\n\n")

	// Create restricted shareable serialized access if share is provided as argument.
	if useAccessShare {
		ShareAccess(access, storjConfig)
	}
}
