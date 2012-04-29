// Copyright 2012 Gustavo Franco <stratus@acm.org>. All rights reserved.
// Use of this source code is governed by a BSD-style license that 
// can be found in the LICENSE file.

package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
)

type Directory struct {
	Scheme  string
	IP      string
	Port    string
	URI     string
	Entries []os.FileInfo
}

var documentRoot = ""
var ip = ""
var ssswsIdentifier = "S3WS 0.1/Beta"

var port = flag.String("port", "8080", "port")
var iface = flag.String("iface", "eth0", "network interface")

func IpByName(iface string) (string, error) {
	ifi, err := net.InterfaceByName(iface)
	if err != nil {
		return "", err
	}
	addrs, err := ifi.Addrs()
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
	servingPath := path.Join(documentRoot, r.URL.Path[1:])
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
	if len(flag.Args()) != 1 {
		log.Fatal("Usage: s3ws [--port=<port>] [--iface=<interface>] DOCUMENTROOT")
	}
	documentRoot = flag.Arg(0)
	http.HandleFunc("/", Serve)
	ip, err := IpByName(*iface)
	if err != nil {
		log.Fatal("IpByName: ", err)
	}
	log.Printf("Serving %s from %s:%s", documentRoot, ip, *port)
	err = http.ListenAndServe(ip+":"+*port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
