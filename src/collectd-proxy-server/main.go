package main

import (
	lib "collectd-proxy-lib"
	"log"
	"net"
	"net/http"
)

// 16 bits counters are used to reference the buffer of messages, so this
// shouldn't be changed without additional changes in the code
const maxMSGS = (1 << 16) - 1

func bufferManager(in chan lib.Msg, out chan lib.Msg, dump chan int) {
	i := uint16(0)
	var buff [maxMSGS]lib.Msg

	for {
		select {
		case m := <-in:
			log.Printf("BM: Msg with %d bytes. Offset %d", len(m), i%maxMSGS)
			buff[i%maxMSGS] = m
			i++
		case <-dump:
			log.Println("BM: Dumping")
			for j := uint16(0); j < i%maxMSGS; j++ {
				out <- buff[j]
			}
			out <- nil
			i = 0
		}
	}
}

func dumpMsgs(w http.ResponseWriter, c chan lib.Msg, dump chan int) {
	dump <- 0
	for m := range c {
		if m == nil {
			break
		}

		header := []byte{uint8(len(m) >> 8), uint8(len(m) & 0xff)}
		_, err := w.Write(append(header, m...))
		if err != nil {
			log.Printf("D: Error: %s", err)
			continue
		}
	}
}

func httpListener(address string, c chan lib.Msg, dump chan int) {
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		dumpMsgs(w, c, dump)
	})
	log.Printf("H: Listening at http://%s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	config, err := lib.GetConfig("/etc/collectd-proxy-server.conf")
	if err != nil {
		log.Fatal(err)
	}

	// comm channels
	storage := make(chan lib.Msg)
	out := make(chan lib.Msg)
	dump := make(chan int)

	// setup buffer manager
	go bufferManager(storage, out, dump)

	// setup http listener
	go httpListener(config.HTTPAddress, out, dump)

	// setup udp listener
	pc, err := net.ListenPacket("udp", config.UDPAddress)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("M: Listening at udp//:%s", pc.LocalAddr())
	for {
		m := make(lib.Msg, 1500)
		n, _, err := pc.ReadFrom(m)
		if err != nil {
			log.Println(err)
			continue
		}
		storage <- m[0:n]
	}
}
