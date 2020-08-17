package main

import (
    "strings"

    "encoding/binary"

)

const headerSize = 12

// Header is the wire format for the DNS packet header.
type Header struct {
    Id                                 uint16
    Bits                               uint16
    Qdcount, Ancount, Nscount, Arcount uint16
}

type Msg struct {
    dh                    *Header
    qname                 string
    rname                 string
    ttl                   int
    question_count        int
    answer_count          int
    authoritative_count   int
    extra_count           int
}

type Error struct {
    err string
}
func (e *Error) Error() string {
    if e == nil {
        return "dns: <nil>"
    }
    return "dns: " + e.err
}

func unpackUint16(msg []byte, off int) (i uint16, off1 int, err error) {
    if off+2 > len(msg) {
        return 0, len(msg), &Error{err: "overflow unpacking uint16"}
    }
    return binary.BigEndian.Uint16(msg[off:]), off + 2, nil
}

func unpackMsgHdr(msg []byte, off int) (Header, int, error) {
    var (
        dh  Header
        err error
    )
    dh.Id, off, err = unpackUint16(msg, off)
    if err != nil {
        return dh, off, err
    }
    dh.Bits, off, err = unpackUint16(msg, off)
    if err != nil {
        return dh, off, err
    }
    dh.Qdcount, off, err = unpackUint16(msg, off)
    if err != nil {
        return dh, off, err
    }
    dh.Ancount, off, err = unpackUint16(msg, off)
    if err != nil {
        return dh, off, err
    }
    dh.Nscount, off, err = unpackUint16(msg, off)
    if err != nil {
        return dh, off, err
    }
    dh.Arcount, off, err = unpackUint16(msg, off)
    if err != nil {
        return dh, off, err
    }
    return dh, off, nil
}

func Uint16ToBytes(n uint16) []byte {
   return []byte{
      byte(n >> 8),
      byte(n),
   }
}

func Uint32ToBytes(n uint32) []byte {
   return []byte{
      byte(n >> 24),
      byte(n >> 16),
      byte(n >> 8),
      byte(n),
   }
}

func (m *Msg)packPTR() ([]byte, int) {
    dhbuf := Uint16ToBytes(m.dh.Id)
    //fmt.Printf("%x\n", dhbuf)
    dhbuf = append(dhbuf[:], []byte("\x85\x00")...)  // 8500 means query success
    // count of Question, Answer, Authoritative, Extra
    dhbuf = append(dhbuf[:], Uint16ToBytes(uint16(m.question_count))...)       //count of Question
    dhbuf = append(dhbuf[:], Uint16ToBytes(uint16(m.answer_count))...)         //count of Answer
    dhbuf = append(dhbuf[:], Uint16ToBytes(uint16(m.authoritative_count))...)  //count of Authoritative
    dhbuf = append(dhbuf[:], Uint16ToBytes(uint16(m.extra_count))...)          //count of Extra

    // Pack Question
    qname_arr := strings.Split(m.qname, ".")
    qnbuf := make([]byte, 0)
    for _, element := range qname_arr {
        // element is the element from someSlice for where we are
        qnbuf = append(qnbuf, uint8(len(element)))
        qnbuf = append(qnbuf, []byte(element)...)
    }
    //dhbuf = append(dhbuf[:], []byte(m.qname)...)    // Qname String
    dhbuf = append(dhbuf[:], qnbuf[:]...)
    dhbuf = append(dhbuf[:], []byte("\x00\x0c")...) // PTR code: 12
    dhbuf = append(dhbuf[:], []byte("\x00\x01")...) // Class IN

    // Pack Answer
    //dhbuf = append(dhbuf[:], []byte(m.qname)...)    // Qname String
    dhbuf = append(dhbuf[:], qnbuf[:]...)
    dhbuf = append(dhbuf[:], []byte("\x00\x0c")...) // PTR code: 12
    dhbuf = append(dhbuf[:], []byte("\x00\x01")...) // Class IN

    rname_arr := strings.Split(m.rname, ".")
    rnbuf := make([]byte, 0)
    for _, element := range rname_arr {
        // element is the element from someSlice for where we are
        rnbuf = append(rnbuf, uint8(len(element)))
        rnbuf = append(rnbuf, []byte(element)...)
    }

    // TTL: 86400 (\x00\x01\x51\x80) for example
    dhbuf = append(dhbuf[:], Uint32ToBytes(uint32(m.ttl))...)
    // Data length
    dhbuf = append(dhbuf[:], Uint16ToBytes(uint16(len(rnbuf)))...)
    dhbuf = append(dhbuf[:], rnbuf[:]...)

    return dhbuf, len(dhbuf)
}
