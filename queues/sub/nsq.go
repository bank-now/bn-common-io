package sub

import (
	"flag"
	"fmt"
	"github.com/bank-now/bn-common-model/common/model"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
)

var (
	showVersion = flag.Bool("version", false, "print version string")

	channel       = flag.String("channel", "default", "NSQ channel")
	maxInFlight   = flag.Int("max-in-flight", 200, "max number of messages to allow in flight")
	totalMessages = flag.Int("n", 0, "total messages to show (will wait if starved)")

	nsqdTCPAddrs     = model.StringArray{}
	lookupdHTTPAddrs = model.StringArray{}
	topics           = model.StringArray{}

	rec Rec

	config Config
)

type Config struct {
	Name    string
	Version string
	Topic   string
	F       Rec
}

type TailHandler struct {
	topicName     string
	totalMessages int
	messagesShown int
}

func (th *TailHandler) HandleMessage(m *nsq.Message) error {
	th.messagesShown++
	rec(m.Body)
	if th.totalMessages > 0 && th.messagesShown >= th.totalMessages {
		os.Exit(0)
	}
	return nil
}

type Rec func(b []byte)

func Setup(c Config) {
	config = c
	cfg := nsq.NewConfig()
	nsqdTCPAddrs = append(nsqdTCPAddrs, "192.168.88.24:4150")
	topics = append(topics, c.Topic)

	flag.Var(&nsq.ConfigFlag{cfg}, "consumer-opt", "http://godoc.org/github.com/nsqio/go-nsq#Config")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s-%s", config.Name, config.Version)
		return
	}

	if *channel == "" {
		rand.Seed(time.Now().UnixNano())
		*channel = fmt.Sprintf("tail%06d#ephemeral", rand.Int()%999999)
	}

	if len(nsqdTCPAddrs) == 0 && len(lookupdHTTPAddrs) == 0 {
		log.Fatal("--nsqd-tcp-address or --lookupd-http-address required")
	}
	if len(nsqdTCPAddrs) > 0 && len(lookupdHTTPAddrs) > 0 {
		log.Fatal("use --nsqd-tcp-address or --lookupd-http-address not both")
	}
	if len(topics) == 0 {
		log.Fatal("--topic required")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Don't ask for more messages than we want
	if *totalMessages > 0 && *totalMessages < *maxInFlight {
		*maxInFlight = *totalMessages
	}

	cfg.UserAgent = fmt.Sprintf("%s-%s", config.Name, config.Version)
	cfg.MaxInFlight = *maxInFlight

	consumers := []*nsq.Consumer{}
	for i := 0; i < len(topics); i += 1 {
		log.Printf("Adding consumer for topic: %s\n", topics[i])

		consumer, err := nsq.NewConsumer(topics[i], *channel, cfg)
		if err != nil {
			log.Fatal(err)
		}

		consumer.AddHandler(&TailHandler{topicName: topics[i], totalMessages: *totalMessages})

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
