// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package hello is a trivial package for gomobile bind example.
package getrelayv2

import (
	//"bufio"
	"io"
	//"bytes"
	"errors"
	"log"
	"net"
	"strings"
	"sync"

	//	"time"
)

var did_base int64 = 0
var state_log = false
var msg_delay_msg = 1000

const (
	recv_header_cap        = 1024 * 10
	recv_buf_cap           = 1024 * 10
	recv_body_cap          = recv_buf_cap
	max_digit_byte         = 1024
	relay_response_timeout = 5 * 1000
	get_relay_retry_time   = 3
)

var (
	did_default    string = "6b326abfca91fe5c9c7fa28612910a6c"
	get_ver_prefix string = "GET /ws/api/getVersion?did="
	http_1_1       string = " HTTP/1.1\r\n"
	host           string = "Host: 0301.dch.dlink.com\r\n"
	alive          string = "Connection: keep-alive\r\n\r\n"
)

/*
POST /connect HTTP/1.1\r\n
Content-Type: text\r\n
\r\n
"did":"1aa2e235723dd6a300381edc551aab3c"\r\n
"hash":"bff84d880edabb2c1598c537906d04d9"\r\n
*/
var (
	connect_prefix string = "POST /connect HTTP/1.1\r\nContent-Type: text\r\nConnection: keep-alive\r\n\r\n"
	connect_did    string = "\"did\":"
	connect_hash   string = "\"hash\":"
)

var api_server string = "r9402.dch.dlink.com:80"

var wg sync.WaitGroup

const max_conn = 10

var connections [max_conn]*net.TCPConn

type ReadCallback interface {
	ReadBytes(data []byte)
}

type Printer interface {
	Print(s string) string
}

func PrintHello(p Printer) string {
	p.Print("Hello, World!")
	return "hello, print"
}

var mycallback ReadCallback

func Getrelay(did string, hash string, callback ReadCallback) (int, error) {
	if did != "" {
		did_default = did
	}
	//get_ver_msg := get_ver_prefix + did_default + http_1_1 + host + alive
	post_msg := "POST /connect HTTP/1.1\r\n\r\n" + "\"did\":\"" + did + "\"" +
		"\r\n" + "\"hash\":\"" + hash + "\"" + "\r\n"
	//connect_msg := connect_prefix + connect_did + "\"" + did + "\"\r\n" + connect_hash + "\"" + hash + "\"\r\n"
	url := api_server

	tcpaddr, err := net.ResolveTCPAddr("tcp", url)
	//tcpaddr, err := net.ResolveTCPAddr("tcp", "52.76.5.88:80")
	log.Println("ResolveTCPAddr = ", tcpaddr.IP)

	if err != nil {
		log.Println("[^mock_app]error", err, " url=", url)
		return -1, err
	}

	conn_retry := 0
	if isConnFull() {
		return -1, errors.New("Connections are Full")
	}
	for conn_retry < 2 { // Retry One time
		conn_retry = conn_retry + 1
		conn, err := net.DialTCP("tcp", nil, tcpaddr) //Connect to server
		if err != nil {
			log.Println("[^mock_app] connect error", err, "url = ", url)
			continue
		}
		//conn.SetKeepAlive(true)
		conn.Write([]byte(post_msg)) //Write to Relay

		newfd := saveConn(conn)
		go readDataRoutine(did, hash, conn, callback)
		return newfd, nil
	}

	return -1, errors.New("TCPDial Error")

}

func readDataRoutine(did string, hash string, conn *net.TCPConn, callback ReadCallback) error {

	first := make([]byte, 1)
	buf := make([]byte, 1024*32)
	log.Println("Start to Read")

	if _, err := conn.Read(first); err == io.EOF {
		log.Println("[^mock_app] Device:", did, " has closed from server")
		conn.Close()
		log.Println("[^mock_app] Fatal !  Exit from program")
	} else {
		for { //continuously read
			_, err := conn.Read(buf)
			if err != nil {
				log.Println("[^mock_app] Device:", did, " Read Error ")
				conn.Close()
				conn = nil
				return errors.New("Read Response Error")
			}
			//log.Println("[^mock_app] Device:", did, " recieve:", string(buf)[:20])
			//isOK := strings.Contains(string(buf), "200 OK")
			is404 := strings.Contains(string(buf), "404 Not Found")
			if is404 {
				log.Println("404 Not Found")
				return errors.New("404 Not Found")
			}

			callback.ReadBytes(buf)
		}

	}
	return nil
}

func saveConn(conn *net.TCPConn) int {
	for i := 0; i < max_conn; i++ {
		if connections[i] == nil {
			connections[i] = conn
			log.Println("[getRelayv2]save connection at ", i)
			return i
		}
	}

	return -1 // Not Found , full of space
}

func isConnFull() bool {
	for i := 0; i < max_conn; i++ {
		if connections[i] == nil {
			return false
		}
	}
	log.Println("[getRelayv2]all connections pool is full")
	return true
}

func Closeall() {
	for i := 0; i < max_conn; i++ {
		if connections[i] != nil {
			connections[i].Close()
		}
	}
	log.Println("[getRelayv2]all connections are closed")
}

func CloseConn(fd int) {
	connections[fd] = nil
	log.Println("[getRelayv2] fd:", fd, " is closed")
}

func replaceConn(fd int, conn *net.TCPConn) int {
	connections[fd] = conn
	return fd
}

func WriteOk(fd int) error {
	var ret error = nil
	if connections[fd] != nil {
		connections[fd].Write([]byte("OK"))
	} else {
		ret = errors.New("fd not found")
	}
	return ret
}
