package main

type Website struct {
	Domain string `json:"domain"`
	HTTP1  bool   `json:"http1"`
	HTTP2  bool   `json:"http2"`
	HTTP3  bool   `json:"http3"`
	Port   uint16 `json:"port"`
}

type Config struct {
	Websites []Website `json:"websites"`
}
