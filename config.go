package main

type Website struct {
	Domain string `json:"domain"`
	Http1  bool   `json:"http1"`
	Http2  bool   `json:"http2"`
	Http3  bool   `json:"http3"`
	Port   uint16 `json:"port"`
}

type Config struct {
	Websites []Website `json:"websites"`
}
