package main

import (
	"log"
	"os"
	"io/ioutil"

	"github.com/miekg/dns"
	"gopkg.in/yaml.v2"
)

var client = new(dns.Client)

func handleDnsRequest(targetUpstream Upstream) func(dns.ResponseWriter, *dns.Msg) {
	log.Printf("Generating handler for %s -> %s", targetUpstream.Target, targetUpstream.At)
	return func(w dns.ResponseWriter, r *dns.Msg) {
		m := r.Copy()
		m.Compress = false

		in, _, err := client.Exchange(m, targetUpstream.At)
		m.SetReply(r)
		if err != nil {
			log.Printf("Query FAIL on (%s -> %s): %s", targetUpstream.Target, targetUpstream.At, err.Error())
			m.Rcode = dns.RcodeServerFailure
			w.WriteMsg(m)
		} else {
			w.WriteMsg(in)
		}
	}
}

type Upstream struct {
	Target	string
	At	string
}
type Conf struct {
	Listen string
	Upstreams []Upstream
}

func main() {
	log.Println("Starting dnsrewrite")

	c := Conf{}
	yamlFile, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Printf("yamlFile err %s", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Printf("unmarshal: %v", err)
	}

	// attach request handlers
	for _, t := range c.Upstreams {
		dns.HandleFunc(t.Target, handleDnsRequest(t))
	}

	// start server
	server := &dns.Server{Addr: c.Listen, Net: "udp"}
	log.Printf("Responder at %s", c.Listen)
	log.Printf("Proxying for %v upstreams", len(c.Upstreams))

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s ", err.Error())
	}
	defer server.Shutdown()
}
