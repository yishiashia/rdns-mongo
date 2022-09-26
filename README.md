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
We run a stress testing with [dnsperf](https://github.com/DNS-OARC/dnsperf) on a 1 cpu, 2G ram VM, and the throughput is **6,000 qps**.

```shell
DNS Performance Testing Tool
Version 2.9.0

[Status] Command line: dnsperf -s 127.0.0.1 -p 5566 -d rdns_query -l 30 -c 20 -Q 10000
[Status] Sending queries (to 127.0.0.1:5566)
[Status] Started at: Mon Sep 26 13:04:52 2022
[Status] Stopping after 30.000000 seconds
[Status] Testing complete (time limit)

Statistics:

  Queries sent:         189957
  Queries completed:    189957 (100.00%)
  Queries lost:         0 (0.00%)

  Response codes:       NOERROR 189957 (100.00%)
  Average packet size:  request 44, response 107
  Run time (s):         30.027059
  Queries per second:   6326.193984

  Average Latency (s):  0.015626 (min 0.000077, max 0.053527)
  Latency StdDev (s):   0.003661
```
