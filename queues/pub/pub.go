package pub

import (
	"fmt"
	"github.com/nsqio/go-nsq"
)

type Config struct {
	Name    string
	Version string
	Address string
	Topic   string
}

func Setup(c Config) (*nsq.Producer, error) {
	cfg := nsq.NewConfig()
	//flag.Var(&nsq.ConfigFlag{cfg}, "producer-opt", "http://godoc.org/github.com/nsqio/go-nsq#Config")
	//flag.Parse()
	cfg.UserAgent = fmt.Sprintf("%s-%s", c.Name, c.Version)

	return nsq.NewProducer(c.Address, cfg)

}
