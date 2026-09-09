package main

type Website struct {
	Http1 bool `json:"http1"`
	Http2 bool `json:"http2"`
	Http3 bool `json:"http3"`
	Port  int  `json:"port"`
}

type Config struct {
	Websites []Website `json:"websites"`
}
