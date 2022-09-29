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

## Performance Testing
We run a stress testing with [dnsperf](https://github.com/DNS-OARC/dnsperf) on a 1 CPU Core, 2 GB RAM virtual machine, and the throughput is about **13,900 qps**.

```shell
DNS Performance Testing Tool
Version 2.9.0

[Status] Command line: dnsperf -s 127.0.0.1 -p 5566 -d rdns_query -l 30 -c 20 -Q 50000
[Status] Sending queries (to 127.0.0.1:5566)
[Status] Started at: Thu Sep 29 22:31:40 2022
[Status] Stopping after 30.000000 seconds
[Status] Testing complete (time limit)

Statistics:

  Queries sent:         417484
  Queries completed:    417484 (100.00%)
  Queries lost:         0 (0.00%)

  Response codes:       NOERROR 417484 (100.00%)
  Average packet size:  request 44, response 107
  Run time (s):         30.012939
  Queries per second:   13910.133893

  Average Latency (s):  0.007103 (min 0.000053, max 0.033680)
  Latency StdDev (s):   0.001920
```
