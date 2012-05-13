// Copyright 2012 Gustavo Franco <stratus@acm.org>. All rights reserved.
// Use of this source code is governed by a BSD-style license that 
// can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
)

import _ "net/http/pprof"

type Directory struct {
	Scheme  string
	IP      string
	Port    string
	URI     string
	Entries []os.FileInfo
}

const ssswsIdentifier = "S3WS 0.2/Beta"

var documentroot = flag.String("documentroot", "", "path to serve")
var ip = ""
var iface = flag.String("iface", "eth0", "network interface")
var port = flag.String("port", "8080", "port")

func IpByName(iface string) (string, error) {
	ifi, err := net.InterfaceByName(iface)
	addrs, err := ifi.Addrs()
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("Error parsing ip address of %s interface.", iface)
	}
	ip = strings.Split(addrs[0].String(), "/")[0]
	return ip, err
}

func Serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", ssswsIdentifier)
	if r.URL.Path == "/favicon.ico" {
		http.NotFound(w, r)
		return
	}
	log.Printf("HTTP GET %s", r.URL.Path)
	servingPath := path.Join(*documentroot, r.URL.Path[1:])
	entries, err := ioutil.ReadDir(servingPath)
	if err != nil {
		log.Println(err)
		log.Printf("Trying to serve %s", servingPath)
		// Failed to open as a dir, not a favicon.ico. Try to serve it.
		http.ServeFile(w, r, servingPath)
	}
	t, _ := template.ParseFiles("s3ws.html")
	t.Execute(w, Directory{"http://", ip, *port, r.URL.Path[1:], entries})
}

func main() {
	flag.Parse()
	if *documentroot == "" {
		log.Fatal("Usage: s3ws --documentroot <documentroot> [--port=<port>] " +
			"[--iface=<interface>]")
	}
	http.HandleFunc("/", Serve)
	ip, err := IpByName(*iface)
	if err != nil {
		log.Fatal("IpByName: ", err)
	}
	log.Printf("Serving %s from %s:%s", *documentroot, ip, *port)
	err = http.ListenAndServe(ip+":"+*port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
