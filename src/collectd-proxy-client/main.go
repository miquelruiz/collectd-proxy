package main

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"net/http"
  lib "collectd-proxy-lib"
)

const endpointHTTP = "http://127.0.0.1:8080/"
const endpointUDP = "127.0.0.1:25826"

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	resp, err := http.Get(endpointHTTP)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	send, err := net.Dial("udp", endpointUDP)
	if err != nil {
		log.Fatalln(err)
	}
	defer send.Close()

	for {
		// read header
		var header [2]byte
		_, err = io.ReadFull(resp.Body, header[:])
		if err != nil {
			if err == io.EOF {
				log.Println("Done")
				break
			}
			log.Fatalln(err)
		}

		// parse header
		size := binary.BigEndian.Uint16(header[:])
		if size == 0 {
			continue
		}

		// read message
		msg := make(lib.Msg, size)
		_, err = io.ReadFull(resp.Body, msg)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Got message of size %d", size)

		// send message
		_, err = send.Write(msg)
		if err != nil {
			log.Println(err)
		}
	}
}
