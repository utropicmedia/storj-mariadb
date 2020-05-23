package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// ConfigMariaDB stores the MariaDB configuration parameters.
type ConfigMariaDB struct {
	HostName   string `json:"hostname"`
	PortNumber string `json:"port"`
	UserName   string `json:"username"`
	Password   string `json:"password"`
	Database   string `json:"database"`
}

// LoadMariaProperty reads and parses the JSON file
// that contains a MariaDB instance's property
// and returns all the properties as an object.
func LoadMariaProperty(fullFileName string) ConfigMariaDB { // fullFileName for fetching database credentials from  given JSON filename.

	var configMariaDB ConfigMariaDB

	// Open and read the file.
	fileHandle, err := os.Open(filepath.Clean(fullFileName))
	if err != nil {
		log.Fatal(err)
	}

	// Decode and parse the JSON properties.
	jsonParser := json.NewDecoder(fileHandle)
	if err = jsonParser.Decode(&configMariaDB); err != nil {
		log.Fatal(err)
	}

	// Close the file handle after reading from it.
	if err = fileHandle.Close(); err != nil {
		log.Fatal(err)
	}

	// Display the read MariaDB configuration properties.
	fmt.Println("\nRead MariaDB configuration from the ", fullFileName, " file")
	fmt.Println("HostName\t", configMariaDB.HostName)
	fmt.Println("PortNumber\t", configMariaDB.PortNumber)
	fmt.Println("UserName \t", configMariaDB.UserName)
	fmt.Println("Password \t", configMariaDB.Password)
	fmt.Println("Database \t", configMariaDB.Database)

	return configMariaDB
}

// CreateBackup creates back-up and returns the corresponding reader object to upload back-up to storj.
func CreateBackup(configMariaDB ConfigMariaDB) *bytes.Reader {

	// Create command to fetch mysqldump from the database.
	cmd := exec.Command("mysqldump", "-P", configMariaDB.PortNumber, "-h", configMariaDB.HostName, "-u", configMariaDB.UserName, "-p"+configMariaDB.Password, configMariaDB.Database)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Create Reader object to use it for uploading data.
	buff := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buff, stdout)
	if err != nil {
		log.Fatal("Error:", err)
	}
	reader := bytes.NewReader(buff.Bytes())

	return reader
}
