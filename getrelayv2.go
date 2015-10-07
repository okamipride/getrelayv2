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
	//"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var did_base int64 = 0
var state_log = false
var msg_delay_msg = 1000

const (
	get_relay_retry_time = 3
	socket_read_timeout  = 2
)

//Error Codes
const (
	error_none                 = 0x00
	error_404                  = 0x01
	error_connect_host         = 0x02
	error_EOF                  = 0x03
	error_ErrNoProgress        = 0x04
	error_ErrShortBuffer       = 0x05
	error_ErrShortWrite        = 0x06
	error_ErrUnexpectedEOF     = 0x07
	error_not_io_error         = 0x08
	error_connect_resolve_host = 0x09
	error_reconnect_fail       = 0x0a
)

var (
	did_default string = "6b326abfca91fe5c9c7fa28612910a6c"
	writeokmsg  string = "HTTP 200 OK\r\nContent-Type: text\r\nContent-Length: 8\r\n\r\nWRITE OK"
)

//messages
var (
	connect_prefix string = "POST /connect HTTP/1.1\r\nContent-Type: text\r\nConnection: keep-alive\r\n\r\n"
	connect_did    string = "\"did\":"
	connect_hash   string = "\"hash\":"
)

//var api_server string = "r9402.dch.dlink.com:80"

var wg sync.WaitGroup

const max_conn = 10

var connections [max_conn]*net.TCPConn
var hosts [max_conn]string

type ReadCallback interface {
	ReadBytes(data []byte)
	RecieveError(err int)
}

var mycallback ReadCallback

func Getrelay(did string, hash string, host string, callback ReadCallback) (int, error) {
	if did != "" {
		did_default = did
	}
	post_msg := "POST /connect HTTP/1.1\r\n\r\n" + "\"did\":\"" + did + "\"" +
		"\r\n" + "\"hash\":\"" + hash + "\"" + "\r\n"

	url := host + ":80"
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

		newfd := saveConn(conn, host)
		go readDataRoutine(newfd, did, hash, conn, callback)
		return newfd, nil
	}
	return -1, errors.New("TCPDial Error")

}

func closeFD(conn *net.TCPConn, fd int, cb ReadCallback, senderr bool, errmsg int) {
	cb.RecieveError(errmsg)
	CloseConn(fd)
}

func reconnect(fd int, did string, hash string, callback ReadCallback) error {
	post_msg := "POST /connect HTTP/1.1\r\n\r\n" + "\"did\":\"" + did + "\"" +
		"\r\n" + "\"hash\":\"" + hash + "\"" + "\r\n"

	var url string
	if hosts[fd] == "" {
		log.Println("getrelayv2]reconnect empty host name=")
		callback.RecieveError(error_connect_host)
		return errors.New("invalid Hostname")
	}

	log.Println("[getrelayv2]hosts[fd]=", hosts[fd])

	url = hosts[fd] + ":80"
	tcpaddr, err := net.ResolveTCPAddr("tcp", url)

	if err != nil {
		log.Println("[getrelayv2] error", err, " url=", url, "fd=", fd)
		callback.RecieveError(error_connect_resolve_host)
		return err
	}

	log.Println("[getrelayv2]reconnect", " fd=", fd, "ResolveTCPAddr = ", tcpaddr)

	conn_retry := 0
	var conn *net.TCPConn
	for conn_retry < 2 { // Retry One time
		conn_retry = conn_retry + 1
		conn, err := net.DialTCP("tcp", nil, tcpaddr) //Connect to server
		if err != nil {
			log.Println("[getrelayv2]connect error", err, "url = ", url)
			continue
		}
		//conn.SetKeepAlive(true)
		conn.Write([]byte(post_msg)) //Write to Relay
		replaceConn(fd, conn, hosts[fd])
		go readDataRoutine(fd, did, hash, conn, callback)
		return nil
	}
	// Shall tell mobile
	closeFD(conn, fd, callback, true, error_connect_host)
	callback.RecieveError(error_reconnect_fail)
	return errors.New("TCPDial Error")
}

func readDataRoutine(fd int, did string, hash string, conn *net.TCPConn, callback ReadCallback) error {
	buf := make([]byte, 1024*32)

	for {
		log.Println("[getrelayv2]fd=", fd, " Start to Read")

		n, err := conn.Read(buf)
		conn.SetReadDeadline(time.Now().Add(socket_read_timeout * time.Second))
		if err != nil || err == io.EOF {
			if err == io.EOF {
				if n > 0 { //Recieve Content
					is404 := strings.Contains(string(buf), "404 Not Found")
					if is404 == true { // Recieve 404 report and close connection, no need to reconnect
						closeFD(conn, fd, callback, true, error_404)
						log.Println("[getrelayv2]readDataRoutine", " did=", did, " fd=", fd, "EOF&404")
						return errors.New("404 found")
					}
					// Not 404,  send to callback and reconnect
					callback.ReadBytes(buf[0:n])
					log.Println("[getrelayv2]readDataRoutine", "sendcallbac:", string(buf[0:n]))
					go reconnect(fd, did, hash, callback)
					log.Println("[getrelayv2]readDataRoutine", " did=", did, " fd=", fd, "EOF&Content")
					return errors.New("EOF found")
				} else { // no content , reconnect again
					//closeFD(conn, fd, callback , false, error_none)
					log.Println("[getrelayv2]readDataRoutine", " did=", did, " fd=", fd, "EOF")
					go reconnect(fd, did, hash, callback)
					return nil
				}
			} else { // Read Error
				closeFD(conn, fd, callback, true, mapIoError(err))
				log.Println("Read Error =", err)
				return errors.New("TCP Read Error:")
			}
		} // End of EOF or read error

		if n > 0 {
			is404 := strings.Contains(string(buf), "404 Not Found")
			if is404 == true { // Recieve 404 report and close connection, no need to reconnect
				closeFD(conn, fd, callback, true, error_404)
				log.Println("[getrelayv2]readDataRoutine", " did=", did, " fd=", fd, "404")
				return errors.New("404 found")
			}
			// Not 404
			callback.ReadBytes(buf[0:n])
			log.Println("[getrelayv2]readDataRoutine", " did=", did, " fd=", fd, " SendToCallback=", string(buf[0:n]))
		}
	}
	return nil
}

func saveConn(conn *net.TCPConn, host string) int {
	for i := 0; i < max_conn; i++ {
		if connections[i] == nil {
			connections[i] = conn
			hosts[i] = host
			log.Println("[getRelayv2]save connection at ", i)
			return i
		}
	}
	return -1 // Not Found , full of space
}

func saveFDConn(fd int, conn *net.TCPConn, host string) {
	connections[fd] = conn
	hosts[fd] = host
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
			hosts[i] = ""
		}
	}
	log.Println("[getRelayv2]all connections are closed")
}

func CloseConn(fd int) {
	if connections[fd] != nil {
		connections[fd].Close()
	}

	connections[fd] = nil
	hosts[fd] = ""
	log.Println("[getRelayv2] fd:", fd, " is closed")
}

func replaceConn(fd int, conn *net.TCPConn, host string) int {
	connections[fd] = conn
	hosts[fd] = host
	return fd
}

func WriteOk(fd int) error {
	var ret error = nil
	if connections[fd] != nil {
		connections[fd].Write([]byte(writeokmsg))
	} else {
		ret = errors.New("fd not found")
	}
	return ret
}

func mapIoError(err error) int {
	if err == io.EOF {
		return error_EOF
	} else if err == io.ErrClosedPipe {
		return error_ErrNoProgress
	} else if err == io.ErrShortBuffer {
		return error_ErrShortBuffer
	} else if err == io.ErrShortWrite {
		return error_ErrShortWrite
	} else if err == io.ErrUnexpectedEOF {
		return error_ErrUnexpectedEOF
	} else {
		return error_not_io_error
	}
}
