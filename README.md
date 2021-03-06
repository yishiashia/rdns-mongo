# rdns-mongo

[![License](https://img.shields.io/badge/license-MIT-blue.svg?maxAge=2592000)](https://opensource.org/licenses/MIT)
[![Golang](https://img.shields.io/badge/Go-v1.15-blue?maxAge=2592000)](https://golang.org/)
[![MongoDB](https://img.shields.io/badge/MongoDB-4.4-green?maxAge=2592000)](https://docs.mongodb.com/manual/)

Golang implemented high currency rdns server with zone data save in mongodb

## Save zone data into mongodb
According to [elgs/dns-zonefile](https://github.com/elgs/dns-zonefile), now we can parse and generate zone file in JSON format.
Therefore, we can easily save it into mongodb.

## Compile and setup
To compile this application, run the following command:
```
$ cd <path-to-workfolder>
$ go build
```

After building application, you can run the application
```
$ ./rdns-mongo
```
