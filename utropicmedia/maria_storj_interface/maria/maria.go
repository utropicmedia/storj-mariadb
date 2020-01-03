// Module to connect to a Maria DB instance
// and fetch its mysqldump.sql.

package maria

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

// DEBUG allows more detailed working to be exposed through the terminal.
var DEBUG = false

// ConfigMariaDB stores the MariaDB configuration parameters.
type ConfigMariaDB struct {
	HostName   string `json:"hostname"`
	PortNumber string `json:"port"`
	UserName   string `json:"username"`
	Password   string `json:"password"`
	Database   string `json:"database"`
}

// MariaReader implements an io.Reader interface.
type MariaReader struct {
	ConfigMariaDB ConfigMariaDB
	lastIndex     int
}

// loadMariaProperty reads and parses the JSON file
// that contain a MariaDB instance's property
// and returns all the properties as an object.
func loadMariaProperty(fullFileName string) (ConfigMariaDB, error) { // fullFileName for fetching database credentials from  given JSON filename.
	var configMariaDB ConfigMariaDB

	// Open and read the file.
	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		return configMariaDB, err
	}
	defer fileHandle.Close()

	// Decode and parse the JSON properties.
	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configMariaDB)

	// Display the read MariaDB configuration properties.
	fmt.Println("Read MariaDB configuration from the ", fullFileName, " file")
	fmt.Println("HostName\t", configMariaDB.HostName)
	fmt.Println("PortNumber\t", configMariaDB.PortNumber)
	fmt.Println("UserName \t", configMariaDB.UserName)
	fmt.Println("Password \t", configMariaDB.Password)
	fmt.Println("Database \t", configMariaDB.Database)

	return configMariaDB, nil
}

// ConnectToDB will connect to a MariaDB instance,
// based on the read property from an external file.
// It returns a reference to an io.Reader with MariaDB instance information
func ConnectToDB(fullFileName string) (*MariaReader, error) { // fullFileName for fetching database credentials from given JSON filename.
	// Read MariaDB instance's properties from an external file.
	configMariaDB, err := loadMariaProperty(fullFileName)

	// Create MariaDB URI
	mariaURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", configMariaDB.UserName, configMariaDB.Password, configMariaDB.HostName, configMariaDB.PortNumber, configMariaDB.Database)

	// Create the database handle, so as to confirm that the connection
	db, err := sql.Open("mysql", mariaURI)
	if err != nil {
		fmt.Println("Error creating MariaDB Client: ", err.Error())
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	// Close the opened database.
	db.Close()

	// Inform about successful connection.
	fmt.Println("Successfully connected to MariaDB!")

	return &MariaReader{ConfigMariaDB: configMariaDB}, nil
}

// Read reads and copies the mysqldump into the buffer.
func (mariaReader *MariaReader) Read(buf []byte) (int, error) { // buf represents the byte array, where data is
	// Create command to fetch mysqldump from the database.
	cmd := exec.Command("mysqldump", "-P", mariaReader.ConfigMariaDB.PortNumber, "-h", mariaReader.ConfigMariaDB.HostName, "-u", mariaReader.ConfigMariaDB.UserName, "-p"+mariaReader.ConfigMariaDB.Password, mariaReader.ConfigMariaDB.Database)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		//
		return 0, err
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
		//
		return 0, err
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal(err)
	}

	if mariaReader.lastIndex < len(bytes) {
		copied := copy(buf[:], bytes[mariaReader.lastIndex:])
		mariaReader.lastIndex = mariaReader.lastIndex + copied
		return len(buf), io.ErrShortBuffer
	}
	return len(buf), io.EOF
}

// FetchData reads ALL tables' data, and
// returns them in appended format.
func FetchData(databaseReader io.Reader) ([]byte, error) { // databaseReader is an io.Reader implementation that 'reads' desired data.
	// Create a buffer of feasible size.
	rawDocumentBSON := make([]byte, 0, 32768)

	// Retrieve ALL tables in the database.
	var allCollectionsDataBSON = []byte{}

	var numOfBytesRead int
	var err error

	// Read data using the given io.Reader.
	for err = io.ErrShortBuffer; err == io.ErrShortBuffer; {
		numOfBytesRead, err = databaseReader.Read(rawDocumentBSON)
		//
		if numOfBytesRead > 0 {
			// Append the read data to earlier one.
			allCollectionsDataBSON = append(allCollectionsDataBSON[:], rawDocumentBSON...)
			//
			if DEBUG {
				fmt.Printf("Read %d bytes of data - Error: %s == %s => %t\n", numOfBytesRead, err, io.ErrShortBuffer, err == io.ErrShortBuffer)
			}
		}
	}
	//
	if DEBUG {
		// complete read data from ALL tables.
		t := time.Now()
		time := t.Format("2006-01-02_15:04:05")
		var filename = "mysqldump_" + time + ".sql"
		err = ioutil.WriteFile(filename, allCollectionsDataBSON, 0644)
	}

	return allCollectionsDataBSON, err
}
