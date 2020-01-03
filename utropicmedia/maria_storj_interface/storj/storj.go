// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package storj

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"storj.io/storj/lib/uplink"
	"storj.io/storj/pkg/macaroon"
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
	SerializedScope      string `json:"serializedScope"`
	DisallowReads        string `json:"disallowReads"`
	DisallowWrites       string `json:"disallowWrites"`
	DisallowDeletes      string `json:"disallowDeletes"`
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
	fmt.Println("Serialized Scope Key\t: ", configStorj.SerializedScope)

	return configStorj, nil
}

// ConnectStorjReadUploadData reads Storj configuration from given file,
// connects to the desired Storj network.
// It then reads data using io.Reader interface and
// uploads it as object to the desired bucket.
func ConnectStorjReadUploadData(fullFileName string, databaseReader io.Reader, databaseName string, keyValue string, restrict string) (string, error) { // fullFileName for fetching storj V3 credentials from  given JSON filename
	// databaseReader is an io.Reader implementation that 'reads' desired data,
	// which is to be uploaded to storj V3 network.
	// databaseName for adding dataBase name in storj V3 filename.
	// Read Storj bucket's configuration from an external file.
	var scope string = ""
	configStorj, err := LoadStorjConfiguration(fullFileName)
	if err != nil {
		log.Fatal("loadStorjConfiguration:", err)
	}

	fmt.Println("\nCreating New Uplink...")

	var cfg uplink.Config
	// Configure the partner id
	cfg.Volatile.PartnerID = "a1ba07a4-e095-4a43-914c-1d56c9ff5afd"

	ctx := context.Background()

	uplinkstorj, err := uplink.NewUplink(ctx, &cfg)
	if err != nil {
		uplinkstorj.Close()
		log.Fatal("Could not create new Uplink object:", err)
	}
	defer uplinkstorj.Close()
	var serializedScope string
	if keyValue == "key" {
		fmt.Println("Parsing the API key...")
		key, err := uplink.ParseAPIKey(configStorj.APIKey)
		if err != nil {
			uplinkstorj.Close()
			log.Fatal("Could not parse API key:", err)
		}

		if DEBUG {
			fmt.Println("API key \t   :", configStorj.APIKey)
			fmt.Println("Serialized API key :", key.Serialize())
		}

		fmt.Println("Opening Project...")
		proj, err := uplinkstorj.OpenProject(ctx, configStorj.Satellite, key)
		if err != nil {
			uplinkstorj.Close()
			//proj.Close()
			log.Fatal("Could not open project:", err)
		}
		defer proj.Close()

		// Creating an encryption key from encryption passphrase.
		if DEBUG {
			fmt.Println("\nGetting encryption key from pass phrase...")
		}

		encryptionKey, err := proj.SaltedKeyFromPassphrase(ctx, configStorj.EncryptionPassphrase)
		if err != nil {
			uplinkstorj.Close()
			proj.Close()
			log.Fatal("Could not create encryption key:", err)
		}

		// Creating an encryption context.
		access := uplink.NewEncryptionAccessWithDefaultKey(*encryptionKey)
		if DEBUG {
			fmt.Println("Encryption access \t:", configStorj.EncryptionPassphrase)
		}

		// Serializing the parsed access, so as to compare with the original key.
		serializedAccess, err := access.Serialize()
		if err != nil {
			uplinkstorj.Close()
			proj.Close()
			log.Fatal("Error Serialized key : ", err)
		}

		if DEBUG {
			fmt.Println("Serialized access key\t:", serializedAccess)
		}

		// Load the existing encryption access context
		accessParse, err := uplink.ParseEncryptionAccess(serializedAccess)
		if err != nil {
			log.Fatal(err)
		}

		if restrict == "restrict" {
			disallowRead, _ := strconv.ParseBool(configStorj.DisallowReads)
			disallowWrite, _ := strconv.ParseBool(configStorj.DisallowWrites)
			disallowDelete, _ := strconv.ParseBool(configStorj.DisallowDeletes)
			userAPIKey, err := key.Restrict(macaroon.Caveat{
				DisallowReads:   disallowRead,
				DisallowWrites:  disallowWrite,
				DisallowDeletes: disallowDelete,
			})
			if err != nil {
				log.Fatal(err)
			}
			userAPIKey, userAccess, err := accessParse.Restrict(userAPIKey,
				uplink.EncryptionRestriction{
					Bucket:     configStorj.Bucket,
					PathPrefix: configStorj.UploadPath,
				},
			)
			if err != nil {
				log.Fatal(err)
			}
			userRestrictScope := &uplink.Scope{
				SatelliteAddr:    configStorj.Satellite,
				APIKey:           userAPIKey,
				EncryptionAccess: userAccess,
			}
			serializedRestrictScope, err := userRestrictScope.Serialize()
			if err != nil {
				log.Fatal(err)
			}
			scope = serializedRestrictScope
			//fmt.Println("Restricted serialized user scope", serializedRestrictScope)
		}
		userScope := &uplink.Scope{
			SatelliteAddr:    configStorj.Satellite,
			APIKey:           key,
			EncryptionAccess: access,
		}
		serializedScope, err = userScope.Serialize()
		if err != nil {
			log.Fatal(err)
		}
		if restrict == "" {
			scope = serializedScope
		}

		proj.Close()
		uplinkstorj.Close()
	} else {
		serializedScope = configStorj.SerializedScope

	}
	parsedScope, err := uplink.ParseScope(serializedScope)
	if err != nil {
		log.Fatal(err)
	}

	uplinkstorj, err = uplink.NewUplink(ctx, &cfg)
	if err != nil {
		log.Fatal("Could not create new Uplink object:", err)
	}
	proj, err := uplinkstorj.OpenProject(ctx, parsedScope.SatelliteAddr, parsedScope.APIKey)
	if err != nil {
		uplinkstorj.Close()
		proj.Close()
		log.Fatal("Could not open project:", err)
	}

	fmt.Println("Opening Bucket: ", configStorj.Bucket)

	// Open up the desired Bucket within the Project.
	bucket, err := proj.OpenBucket(ctx, configStorj.Bucket, parsedScope.EncryptionAccess)
	if err != nil {
		fmt.Println("Could not open bucket", configStorj.Bucket, ":", err)
		fmt.Println("Trying to create new bucket....")
		_, err1 := proj.CreateBucket(ctx, configStorj.Bucket, nil)
		if err1 != nil {
			uplinkstorj.Close()
			proj.Close()
			bucket.Close()
			fmt.Printf("Could not create bucket %q:", configStorj.Bucket)
			log.Fatal(err1)
		} else {
			fmt.Println("Created Bucket", configStorj.Bucket)
		}
		fmt.Println("Opening created Bucket: ", configStorj.Bucket)
		bucket, err = proj.OpenBucket(ctx, configStorj.Bucket, parsedScope.EncryptionAccess)
		if err != nil {
			fmt.Printf("Could not open bucket %q:", configStorj.Bucket)
			log.Fatal(err)
		}
	}

	defer bucket.Close()


	var fileNamesDEBUG []string
	checkSlash := configStorj.UploadPath[len(configStorj.UploadPath)-1:]
	if checkSlash != "/" {
		configStorj.UploadPath = configStorj.UploadPath + "/"
	}

	// Read data using io.Reader and upload it to Storj.
	for err = io.ErrShortBuffer; err == io.ErrShortBuffer; {
		t := time.Now()
		timeNow := t.Format("2006-01-02_15:04:05")
		var filename = databaseName + "/mysqldump_" + timeNow + ".sql"
		//
		fmt.Println("File path: ", configStorj.UploadPath+filename)
		fmt.Println("\nUploading of the object to the Storj bucket: Initiated...")

		err = bucket.UploadObject(ctx, configStorj.UploadPath+filename, databaseReader, nil)

		if DEBUG {
			fileNamesDEBUG = append(fileNamesDEBUG, filename)
		}
	}

	if err != nil {
		fmt.Printf("Could not upload: %s\t", err)
		return scope, err
	}

	fmt.Println("Uploading of the object to the Storj bucket: Completed!")

	if DEBUG {
		for _, filename := range fileNamesDEBUG {
			// Test uploaded data by downloading it.
			// serializedAccess, err := access.Serialize().
			// Initiate a download of the same object again.
			readBack, err := bucket.OpenObject(ctx, configStorj.UploadPath+filename)
			if err != nil {
				return scope, fmt.Errorf("could not open object at %q: %v", configStorj.UploadPath+filename, err)
			}
			defer readBack.Close()

			fmt.Println("\nDownloading range")
			// We want the whole thing, so range from 0 to -1.
			strm, err := readBack.DownloadRange(ctx, 0, -1)
			if err != nil {
				return scope, fmt.Errorf("could not initiate download: %v", err)
			}
			defer strm.Close()

			downloadFileName := strings.Split(filename, ":")
			path := strings.Split(downloadFileName[0], "/")

			var fileNameDownload = "./debug/" + downloadFileName[0] + "_" + downloadFileName[1] + "_" + downloadFileName[2]
			// Create directory if not present
			if _, err := os.Stat("./debug"); os.IsNotExist(err) {
				err1 := os.Mkdir("./debug", os.ModeDir)
				if err1 != nil {
					log.Fatal("Invalid Download Path")
				}
			}

			if _, err := os.Stat("./debug/" + path[0]); os.IsNotExist(err) {
				err1 := os.Mkdir("./debug/"+path[0], os.ModeDir)
				if err1 != nil {
					log.Fatal("Invalid Download Path")
				}
			}

			fmt.Printf("Downloading Object %s from bucket : Initiated...\n", downloadFileName[0]+"_"+downloadFileName[1]+"_"+downloadFileName[2])
			// Read everything from the stream.
			receivedContents, err := ioutil.ReadAll(strm)
			if err != nil {
				return scope, fmt.Errorf("could not read object: %v", err)
			}

			err = ioutil.WriteFile(fileNameDownload, receivedContents, 0644)
			if err != nil {
				fmt.Println(err)
			}

			fmt.Printf("Downloaded %d bytes of Object from bucket!\n", len(receivedContents))
		}
	}

	return scope, nil
}
