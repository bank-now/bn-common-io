package sub

import (
	"flag"
	"fmt"
	"github.com/bank-now/bn-common-model/common/model"
	"github.com/nsqio/go-nsq"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	showVersion = flag.Bool("version", false, "print version string")

	channel       = flag.String("channel", "default", "NSQ channel")
	maxInFlight   = flag.Int("max-in-flight", 200, "max number of messages to allow in flight")
	totalMessages = flag.Int("n", 0, "total messages to show (will wait if starved)")

	nsqdTCPAddrs     = model.StringArray{}
	lookupdHTTPAddrs = model.StringArray{}
	topics           = model.StringArray{}
)

type Config struct {
	Name    string
	Version string
	Address string
	Topic   string
	F       ReceiveFunction
}

type tailHandler struct {
	topicName     string
	totalMessages int
	messagesShown int
	config        Config
}

func (th *tailHandler) HandleMessage(m *nsq.Message) error {
	th.messagesShown++
	th.config.F(m.Body)
	if th.totalMessages > 0 && th.messagesShown >= th.totalMessages {
		os.Exit(0)
	}
	return nil
}

type ReceiveFunction func(b []byte)

func Subscribe(c Config) {
	cfg := nsq.NewConfig()
	nsqdTCPAddrs = append(nsqdTCPAddrs, c.Address)
	topics = append(topics, c.Topic)

	flag.Var(&nsq.ConfigFlag{cfg}, "consumer-opt", "http://godoc.org/github.com/nsqio/go-nsq#Config")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s-%s", c.Name, c.Version)
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Don't ask for more messages than we want
	if *totalMessages > 0 && *totalMessages < *maxInFlight {
		*maxInFlight = *totalMessages
	}

	cfg.UserAgent = fmt.Sprintf("%s-%s", c.Name, c.Version)
	cfg.MaxInFlight = *maxInFlight

	consumers := []*nsq.Consumer{}
	for i := 0; i < len(topics); i += 1 {
		log.Printf("Adding consumer for topic: %s\n", topics[i])

		consumer, err := nsq.NewConsumer(topics[i], *channel, cfg)
		if err != nil {
			log.Fatal(err)
		}

		consumer.AddHandler(&tailHandler{topicName: topics[i], totalMessages: *totalMessages, config: c})

		err = consumer.ConnectToNSQDs(nsqdTCPAddrs)
		if err != nil {
			log.Fatal(err)
		}

		err = consumer.ConnectToNSQLookupds(lookupdHTTPAddrs)
		if err != nil {
			log.Fatal(err)
		}

		consumers = append(consumers, consumer)
	}

	<-sigChan

	for _, consumer := range consumers {
		consumer.Stop()
	}
	for _, consumer := range consumers {
		<-consumer.StopChan
	}
}
