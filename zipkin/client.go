package main

import (
	"encoding/json"
	"fmt"
	"github.com/bank-now/bn-common-io/rest"
	"github.com/bank-now/bn-common-io/util"
	"github.com/golang-plus/uuid"

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
func NewChildSpan(parent Span, methodName string) Span {
	s := Span{
		ID:        util.RandomHexString(LenID),
		ParentId:  parent.ID,
		TraceId:   parent.TraceId,
		Name:      methodName,
		Timestamp: time.Now().UnixNano() / int64(1000),
		LocalEndpoint: LocalEndpoint{
			ServiceName: parent.LocalEndpoint.ServiceName,
		},
	}
	return s
}

func Send(url string, s []Span) (body []byte, err error) {
	b, err := json.Marshal(s)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))
	body, err = rest.Post(url, b)
	return
}

func main() {

	var spans []Span

	parent := NewSpan("100.test.com", "createPenguin")
	parent.Duration = 100000
	spans = append(spans, parent)
	time.Sleep(100 * time.Millisecond)

	child := NewChildSpan(parent, "washPenguin")
	child.Duration = 200000
	spans = append(spans, child)

	body, err := Send("http://192.168.88.24:9411/api/v2/spans", spans)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(body))
	}

}
