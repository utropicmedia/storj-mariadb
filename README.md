# storj-mariadb connector

## Initial Set-up
Make sure your `PATH` includes the `$GOPATH/bin` directory, so that your commands can be easily used [Refer: Install the Go Tools](https://golang.org/doc/install):
```
export PATH=$PATH:$GOPATH/bin
```

Install [github.com/urfave/cli](https://github.com/urfave/cli), by running:
```
$ go get github.com/urfave/cli
```

Install [go-mysql-driver](https://github.com/go-sql-driver/mysql) go package, by running:
```
$ go get -u github.com/go-sql-driver/mysql
```

Install [storj-uplink](https://godoc.org/storj.io/storj/lib/uplink) go package, by running:
```
$ go get storj.io/storj/lib/uplink
```



## Configure Packages
```
$ chmod 555 configure.sh
$ ./configure.sh
```

**NOTE**: In Windows powershell, the corresponding command is:      
```     
> sh ./configure.sh     
```

## Build ONCE
```
$ go build storj-mariadb.go
```


## Set-up Files
* Create a `db_property.json` file, with following contents about a MariaDB instance:
```json
    { 
        "hostname": "hostName",
        "port":     "27017",
        "username": "userName",
        "password": "password",
        "database": "databaseName"
    }
```

* Create a `storj_config.json` file, with Storj network's configuration information in JSON format:
```json
    { 
        "apikey":     "change-me-to-the-api-key-created-in-satellite-gui",
        "satellite":  "us-central-1.tardigrade.io:7777",
        "bucket":     "my-first-bucket",
        "uploadPath": "foo/bar/baz",
        "encryptionpassphrase": "test"
    }
```

* Store both these files in a `config` folder.  Filename command-line arguments are optional.  defualt locations are used.


## Run the command-line tool

**NOTE**: The following commands operate in a Linux system

* Get help
```
    $ ./storj-mariadb -h
```

* Check version
```
    $ ./storj-mariadb -v
```

* Read mysqldump from desired MariaDB instance and upload it to given Storj network bucket.  [note: filename arguments are optional.  default locations are used.]
```
    $ ./storj-mariadb store ./config/db_property.json ./config/storj_config.json  
```

* Read mysqldump from desired MariaDB instance in `debug` mode and upload it to given Storj network bucket.  [note: filename arguments are optional.  default locations are used.]
```
    $ ./storj-mariadb store debug ./config/db_property.json ./config/storj_config.json  
```

* Read MariaDB instance property from a desired JSON file and fetch its mysqldump
```
    $ ./storj-mariadb parse   
```

* Read MariaDB instance property in `debug` mode from a desired JSON file and fetch its mysqldump
```
    $ ./storj-mariadb.go parse debug 
```

* Read and parse Storj network's configuration, in JSON format, from a desired file and upload a sample object
```
    $ ./storj-mariadb.go test 
```
* Read and parse Storj network's configuration, in JSON format, from a desired file and upload a sample object in `debug` mode
```
    $ ./storj-mariadb.go test debug 
```
