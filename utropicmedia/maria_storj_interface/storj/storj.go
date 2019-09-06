// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package storj

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"storj.io/storj/lib/uplink"
)

// DEBUG allows more detailed working to be exposed through the terminal.
var DEBUG = false

// ConfigStorj depicts keys to search for within the stroj_config.json file.
type ConfigStorj struct {
	APIKey               string `json:"apikey"`
	Satellite            string `json:"satellite"`
	Bucket               string `json:"bucket"`
	UploadPath           string `json:"uploadPath"`
	EncryptionPassphrase string `json:"encryptionpassphrase"`
}

// LoadStorjConfiguration reads and parses the JSON file that contain Storj configuration information.
func LoadStorjConfiguration(fullFileName string) (ConfigStorj, error) { // fullFileName for fetching storj V3 credentials from  given JSON filename.

	var configStorj ConfigStorj

	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		return configStorj, err
	}
	defer fileHandle.Close()

	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configStorj)

	// Display read information.
	fmt.Println("\nRead Storj configuration from the ", fullFileName, " file")
	fmt.Println("\nAPI Key\t\t: ", configStorj.APIKey)
	fmt.Println("Satellite	: ", configStorj.Satellite)
	fmt.Println("Bucket		: ", configStorj.Bucket)
	fmt.Println("Upload Path\t: ", configStorj.UploadPath)

	return configStorj, nil
}

// ConnectStorjReadUploadData reads Storj configuration from given file,
// connects to the desired Storj network.
// It then reads data using io.Reader interface and
// uploads it as object to the desired bucket.
func ConnectStorjReadUploadData(fullFileName string, databaseReader io.Reader, databaseName string) error { // fullFileName for fetching storj V3 credentials from  given JSON filename
	// databaseReader is an io.Reader implementation that 'reads' desired data,
	// which is to be uploaded to storj V3 network.
	// databaseName for adding dataBase name in storj V3 filename.
	// Read Storj bucket's configuration from an external file.
	configStorj, err := LoadStorjConfiguration(fullFileName)
	if err != nil {
		return fmt.Errorf("loadStorjConfiguration: %s", err)
	}

	fmt.Println("\nCreating New Uplink...")

	var cfg uplink.Config
	// configure the partner id
	cfg.Volatile.PartnerID = "a1ba07a4-e095-4a43-914c-1d56c9ff5afd"

	ctx := context.Background()

	uplinkstorj, err := uplink.NewUplink(ctx, &cfg)
	if err != nil {
		return fmt.Errorf("Could not create new Uplink object: %s", err)
	}
	defer uplinkstorj.Close()

	fmt.Println("Parsing the API key...")
	key, err := uplink.ParseAPIKey(configStorj.APIKey)
	if err != nil {
		return fmt.Errorf("Could not parse API key: %s", err)
	}

	if DEBUG {
		fmt.Println("API key \t   :", key)
		fmt.Println("Serialized API key :", key.Serialize())
	}

	fmt.Println("Opening Project...")
	proj, err := uplinkstorj.OpenProject(ctx, configStorj.Satellite, key)

	if err != nil {
		return fmt.Errorf("Could not open project: %s", err)
	}
	defer proj.Close()

	// Creating an encryption key from encryption passphrase.
	if DEBUG {
		fmt.Println("\nGetting encryption key from pass phrase...")
	}

	encryptionKey, err := proj.SaltedKeyFromPassphrase(ctx, configStorj.EncryptionPassphrase)
	if err != nil {
		return fmt.Errorf("Could not create encryption key: %s", err)
	}

	// Creating an encryption context.
	access := uplink.NewEncryptionAccessWithDefaultKey(*encryptionKey)
	fmt.Println("Encryption access \t:", *access)

	// Serializing the parsed access, so as to compare with the original key.
	serializedAccess, err := access.Serialize()
	if err != nil {
		fmt.Println("Error Serialized key : ", err)
	}

	if DEBUG {
		fmt.Println("Serialized access key\t:", serializedAccess)
	}
	fmt.Println("Opening Bucket: ", configStorj.Bucket)

	// Open up the desired Bucket within the Project.
	bucket, err := proj.OpenBucket(ctx, configStorj.Bucket, access)
	//
	if err != nil {
		return fmt.Errorf("Could not open bucket %q: %s", configStorj.Bucket, err)
	}
	defer bucket.Close()

	//fmt.Println("Getting data into a buffer...")
	
	var fileNamesDEBUG []string
	
	// Read data using io.Reader and upload it to Storj.
	for err = io.ErrShortBuffer; (err == io.ErrShortBuffer); {
		t := time.Now()
		timeNow := t.Format("2006-01-02_15:04:05")
		var filename = databaseName + "/mysqldump_" + timeNow + ".sql"
		//
		fmt.Println("File path: ", configStorj.UploadPath + filename)
		fmt.Println("\nUploading of the object to the Storj bucket: Initiated...")
		
		err = bucket.UploadObject(ctx, configStorj.UploadPath + filename, databaseReader, nil)
		//
		if DEBUG {
			fileNamesDEBUG = append(fileNamesDEBUG, filename)
			//
			fmt.Printf("\nbucket.UploadObject - Error: %s == %s => %t\n", err, io.ErrShortBuffer, err == io.ErrShortBuffer)
		}
	}
	
	if err != nil {
		return fmt.Errorf("Could not upload: %s", err)
	}

	fmt.Println("Uploading of the object to the Storj bucket: Completed!")

	if DEBUG {
		for _, filename := range fileNamesDEBUG {
			// Test uploaded data by downloading it.
			// serializedAccess, err := access.Serialize().
			// Initiate a download of the same object again.
			readBack, err := bucket.OpenObject(ctx, configStorj.UploadPath + filename)
			if err != nil {
				return fmt.Errorf("could not open object at %q: %v", configStorj.UploadPath + filename, err)
			}
			defer readBack.Close()

			fmt.Println("\nDownloading range")
			// We want the whole thing, so range from 0 to -1.
			strm, err := readBack.DownloadRange(ctx, 0, -1)
			if err != nil {
				return fmt.Errorf("could not initiate download: %v", err)
			}
			defer strm.Close()
			fmt.Printf("Downloading Object %s from bucket : Initiated...\n", filename)
			// Read everything from the stream.
			receivedContents, err := ioutil.ReadAll(strm)
			if err != nil {
				return fmt.Errorf("could not read object: %v", err)
			}
			var fileNameDownload = "downloadeddata/" + filename + ".sql"
			err = ioutil.WriteFile(fileNameDownload, receivedContents, 0644)

			fmt.Printf("Downloaded %d bytes of Object from bucket!\n", len(receivedContents))
		}
	}

	return nil
}
