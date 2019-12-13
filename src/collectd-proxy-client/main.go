package main

import (
	lib "github.com/miquelruiz/collectd-proxy/src/collectd-proxy-lib"
	"encoding/binary"
	"io"
	"log"
	"net"
	"net/http"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	config, err := lib.GetConfig("/etc/collectd-proxy-client.conf")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Get(config.HTTPAddress)
	if err != nil {
		log.Fatalf("Failed to connect to '%s': %s", config.HTTPAddress, err)
	}
	defer resp.Body.Close()

	send, err := net.Dial("udp", config.UDPAddress)
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
