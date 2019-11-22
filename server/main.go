package main

import (
	"flag"
	"os"

	gohttp "net/http"

	"github.com/ian-kent/go-log/log"
	"github.com/koshatul/mailhog/common/http"
	comcfg "github.com/koshatul/mailhog/core/config"
	"github.com/koshatul/mailhog/server/api"
	"github.com/koshatul/mailhog/server/config"
	"github.com/koshatul/mailhog/server/smtp"
	"github.com/koshatul/mailhog/ui/assets"
)

var conf *config.Config
var comconf *comcfg.Config
var exitCh chan int

func configure() {
	comcfg.RegisterFlags()
	config.RegisterFlags()
	flag.Parse()
	conf = config.Configure()
	comconf = comcfg.Configure()
}

func main() {
	configure()

	if comconf.AuthFile != "" {
		http.AuthFile(comconf.AuthFile)
	}

	exitCh = make(chan int)
	cb := func(r gohttp.Handler) {
		api.CreateAPI(conf, r)
	}
	go http.Listen(conf.APIBindAddr, assets.Asset, exitCh, cb)
	go smtp.Listen(conf, exitCh)

	for {
		select {
		case <-exitCh:
			log.Printf("Received exit signal")
			os.Exit(0)
		}
	}
}
