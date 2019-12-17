package main

import (
	"log"
	"strings"
	"os"

	"github.com/miekg/dns"
)

var client = new(dns.Client)

func Query(data string) []dns.RR {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(data), dns.TypeA)

	in, _, err := client.Exchange(m, os.Args[2])
	if err == nil {
		if len(in.Answer) > 0 {
			return in.Answer
		}
	} else {
		log.Printf("Query FAIL: %s", err.Error())
	}
	return nil
}

func parseQuery(m *dns.Msg, w dns.ResponseWriter) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			log.Printf("Query for A %s", q.Name)
			if strings.HasSuffix(q.Name, dns.Fqdn(os.Args[3])) {
				rr := Query(q.Name)
				if len(rr) > 0 {
					log.Printf("Valid, resolved to %s", rr)
					m.Answer = rr
				} else {
					log.Printf("Did not resolve")
					m.Rcode = dns.RcodeNameError
				}
			}
		}
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m, w)
	}

	w.WriteMsg(m)
}

func main() {
	// attach request handler
	dns.HandleFunc(".", handleDnsRequest)

	// start server
	server := &dns.Server{Addr: os.Args[1], Net: "udp"}
	log.Printf("Responder at %s", os.Args[1])
	log.Printf("Upstream at %s", os.Args[2])
	log.Printf("Filter for %s", os.Args[3])
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s ", err.Error())
	}
}
