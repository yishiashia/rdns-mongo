package main

import (
    "fmt"
    "net"
    "log"
    "strings"
    "context"
    "runtime"
    "time"
    "sync"
    "sync/atomic"
    "os"
    "os/signal"
    "strconv"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "github.com/libp2p/go-reuseport"
)

func InitiateMongoClient() (*mongo.Client, context.Context) {
    var err error
    var client *mongo.Client
    credential := options.Credential{
        Username: "<YOUR_MONGODB_ACCT>",
        Password: "<YOUR_MONGODB_PASSWORD>",
    }

    ctx := context.Background()

    uri := "mongodb://localhost:27017"
    opts := options.Client()
    opts.ApplyURI(uri)
    opts.SetAuth(credential)
    opts.SetMaxPoolSize(100)
    if client, err = mongo.Connect(ctx, opts); err != nil {
        fmt.Println(err.Error())
    }
    return client, ctx
}

type Message struct {
    recipient *net.UDPAddr
    data      []byte
    length    int
}

type messageQueue chan Message

const (
    maxQueueSize  = 1000000
    UDPPacketSize = 1500
)

var ops uint64 = 0
var total uint64 = 0
var inmq = make(messageQueue, maxQueueSize)
var outmq = make(messageQueue, maxQueueSize)
var bufferPool sync.Pool

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    bufferPool = sync.Pool{
        New: func() interface{} { return make([]byte, UDPPacketSize) },
    }

    beginListen()

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func() {
        for range c {
            atomic.AddUint64(&total, ops)
            log.Printf("Total ops %d", total)
            os.Exit(0)
        }
    }()

    flushInterval := time.Duration(1) * time.Second
    flushTicker := time.NewTicker(flushInterval)
    for range flushTicker.C {
        //log.Printf("Ops/s %f", float64(ops)/flushInterval.Seconds())
        atomic.AddUint64(&total, ops)
        atomic.StoreUint64(&ops, 0)
    }
}

func beginListen() {
    address := "0.0.0.0:53"
    conn, err := reuseport.ListenPacket("udp", address)
    //defer conn.Close()

    fmt.Printf("UDP server start and listening on %s.\n", address)

    if err != nil {
       panic(err)
    }

    // Connect to MongoDB
    client, _ := InitiateMongoClient()

    for i:=0; i<runtime.NumCPU(); i++ {
        go outmq.sendFromOutbox(conn.(*net.UDPConn))
        go inmq.dequeue(client)
        go serve(conn.(*net.UDPConn))
    }
}


func serve(conn *net.UDPConn) {
    for {
        buf := bufferPool.Get().([]byte)

        n, addr, err := conn.ReadFromUDP(buf)
        if err != nil {
            fmt.Println("failed read udp msg, error: " + err.Error())
            continue
        }

    inmq <- Message{recipient: addr, data: buf, length: n}
    }
}

func (mq messageQueue) sendFromOutbox(conn *net.UDPConn) {
    n, err := 0, error(nil)
    for msg := range mq {
        n, err = conn.WriteToUDP(msg.data, msg.recipient)
        if err != nil {
            panic(err)
        }
        if n != len(msg.data) {
            log.Println("Tried to send", len(msg.data), "bytes but only sent ", n)
        }
        atomic.AddUint64(&ops, 1)
    }
}

func (mq messageQueue) dequeue(client *mongo.Client) {
    for m := range mq {
        processRequest(m.recipient, m.data[0:m.length], m.length, client)
        bufferPool.Put(m.data)
    }
}

func processRequest(addr *net.UDPAddr, buf []byte, n int, client *mongo.Client) {
    dh, off, err := unpackMsgHdr(buf[:n], 0)

    if err != nil {
        return
    }

    output_msg := new(Msg)
    output_msg.dh = &dh
    output_msg.question_count = 1
    output_msg.answer_count = 1
    output_msg.authoritative_count = 0
    output_msg.extra_count = 0

    off1 := off
    maxDomainNamePresentationLength := 61*4 + 1 + 63*4 + 1 + 63*4 + 1 + 63*4 + 1
    s := make([]byte, 0, maxDomainNamePresentationLength)
    for {
        c := int(buf[off1])
        s = append(s, buf[off1])
        off1++

        if c == 0 {
            // end of name
            break
        }
    }

    var q_arr []string
    idx := 0
    for {
        byte_len := int(s[idx])
        if s[idx] == 0x00 {
            break
        } else {
            q_arr = append(q_arr, string(s[idx+1:idx+byte_len+1]))
            idx += byte_len + 1
            if idx >= len(s) {
                break
            }
        }
    }
    qname := strings.Join(q_arr, ".")
    output_msg.qname = qname + "."

    // Query MongoDB to get PTR record
    if len(q_arr) > 1 {
        query_key := strings.Join(q_arr[1:], ".")
        //fmt.Printf("%s\n", query_key)
        // Connect to MongoDB
        if err != nil {
            fmt.Println("failed read udp msg, error: " + err.Error())
            return
        }

        output_msg.rname = ""
        if len(q_arr) > 4 {
            output_msg.rname = fmt.Sprintf("%s-%s-%s-%s.example.com.", q_arr[3], q_arr[2], q_arr[1], q_arr[0])
        }

        ctx, cancel := context.WithTimeout(context.Background(), 6 * time.Second)
        defer cancel()
        collection := client.Database("dns").Collection("reverse")

        filter := bson.M{"origin": query_key}
        res, err := collection.FindOne(ctx, filter).DecodeBytes()
        if err != nil {
            //log.Fatalf("FindOne error: %v", err)
            fmt.Printf("FindOne error: %v, %s\n", err, query_key)
        } else {
            var zonedata primitive.M
            bson.Unmarshal(res, &zonedata)
            //fmt.Println(zonedata["ptr"])
            //fmt.Printf("%T\n", zonedata["ptr"])
            //fmt.Println("type: ", reflect.TypeOf(zonedata["ptr"]))

            ptrs := zonedata["ptr"].(primitive.A)
            ttl, err := strconv.Atoi(zonedata["ttl"].(string))
            for _, e := range ptrs {
                m := e.(primitive.M)
                //fmt.Println(m["fullname"].(string))
                if strings.HasPrefix(m["fullname"].(string), qname) {
                    output_msg.rname = m["host"].(string)
                    if err != nil {
                        output_msg.ttl = 86400
                    } else {
                        output_msg.ttl = ttl
                    }
                    break
                }
            }
        }

    od, l := output_msg.packPTR()
    msg := Message{recipient: addr, data: od, length: l}
        outmq <- msg
    }
}
