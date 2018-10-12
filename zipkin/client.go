package zipkin

import (
	"encoding/json"
	"fmt"
	"github.com/bank-now/bn-common-io/queues/pub"
	"github.com/bank-now/bn-common-io/util"
	"github.com/golang-plus/uuid"
	"github.com/nsqio/go-nsq"

	"time"
)

const (
	LenID = 16
)

type Span struct {
	TraceId       string        `json:"traceId"`
	ParentId      string        `json:"parentId,omitempty"`
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Timestamp     int64         `json:"timestamp"`
	Duration      int64         `json:"duration,omitempty"`
	LocalEndpoint LocalEndpoint `json:"localEndpoint"`
}

type Ghost struct {
	TraceId     string `json:"traceId"`
	ID          string `json:"ID,omitempty"`
	ServiceName string `json:"serviceName"`
}

func (parent *Span) ToGhost() Ghost {
	return Ghost{
		ID:          parent.ID,
		TraceId:     parent.TraceId,
		ServiceName: parent.LocalEndpoint.ServiceName}
}

type LocalEndpoint struct {
	ServiceName string `json:"serviceName"`
}

func NewSpan(serviceName, methodName string) Span {
	u, _ := uuid.NewRandom()
	id := util.RandomHexString(LenID)
	s := Span{
		ID:        id,
		TraceId:   u.Format(uuid.StyleWithoutDash),
		Name:      methodName,
		Timestamp: time.Now().UnixNano() / int64(1000),
		LocalEndpoint: LocalEndpoint{
			ServiceName: serviceName,
		},
	}
	return s
}
func NewChildSpan(parent Ghost, serviceName string, methodName string) Span {
	s := Span{
		ID:        util.RandomHexString(LenID),
		ParentId:  parent.ID,
		TraceId:   parent.TraceId,
		Name:      methodName,
		Timestamp: time.Now().UnixNano() / int64(1000),
		LocalEndpoint: LocalEndpoint{
			ServiceName: serviceName,
		},
	}
	return s
}

func EnqueueSpan(producer *nsq.Producer, config pub.Config, s []Span) (body []byte, err error) {

	b, err := json.Marshal(s)
	err = producer.Publish(config.Topic, b)
	if err != nil {
		fmt.Println(err)
		return
	}
	//body, err = rest.Post(url, b)
	return
}

func LogParentSpan(producer *nsq.Producer, config pub.Config, s Span, d time.Duration) (ghost Ghost, err error) {
	var spans []Span
	s.Duration = d.Nanoseconds() / 1000
	spans = append(spans, s)
	_, err = EnqueueSpan(producer, config, spans)
	if err != nil {
		return
	}
	ghost = s.ToGhost()
	return
}

func LogParent(producer *nsq.Producer, config pub.Config, serviceName, methodName string, d time.Duration) (ghost Ghost, err error) {
	s := NewSpan(serviceName, methodName)
	return LogParentSpan(producer, config, s, d)
}

func LogChild(producer *nsq.Producer, config pub.Config, parent Ghost, serviceName string, methodName string, d time.Duration) Ghost {
	var spans []Span
	s := NewChildSpan(parent, serviceName, methodName)
	s.Duration = d.Nanoseconds() / 1000
	spans = append(spans, s)
	_, err := EnqueueSpan(producer, config, spans)
	if err != nil {
		fmt.Println(err)
	}
	return s.ToGhost()

}
