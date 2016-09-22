package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hprose/hprose-go/hprose"
	"github.com/toolkits/logger"

	"github.com/coraldane/DelayQ/g"
	"github.com/coraldane/DelayQ/src"
)

func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")

	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	g.ParseConfig(*cfg)
	log.Println(g.Config())

	logger.SetLevelWithDefault(g.Config().LogLevel, "info")
	g.InitRedisConnPool()

	go src.Schedule()

	var err error
	if "tcp" == g.Config().Protocol {
		err = startTcpServer()
	} else {
		err = startHttpServer()
	}
	if nil != err {
		log.Println("start server error:", err)
	}
}

func startHttpServer() error {
	server := hprose.NewHttpService()
	server.AddMethods(new(src.DelayQueue))
	return http.ListenAndServe(g.Config().Listen, server)
}

func startTcpServer() error {
	server := hprose.NewTcpServer(fmt.Sprintf("tcp://%s/", g.Config().Listen))
	server.AddMethods(new(src.DelayQueue))
	return server.Start()
}
