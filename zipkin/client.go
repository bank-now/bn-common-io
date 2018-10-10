package zipkin

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
func NewChildSpan(parent Ghost, methodName string) Span {
	s := Span{
		ID:        util.RandomHexString(LenID),
		ParentId:  parent.ID,
		TraceId:   parent.TraceId,
		Name:      methodName,
		Timestamp: time.Now().UnixNano() / int64(1000),
		LocalEndpoint: LocalEndpoint{
			ServiceName: parent.ServiceName,
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
	body, err = rest.Post(url, b)
	return
}

func LogParentFromSpan(url string, s Span, d int64) Ghost {
	var spans []Span
	s.Duration = d
	spans = append(spans, s)
	_, err := Send(url, spans)
	if err != nil {
		fmt.Println(err)
	}
	return s.ToGhost()
}

func LogParent(url, serviceName, methodName string, d int64) Ghost {
	s := NewSpan(serviceName, methodName)
	return LogParentFromSpan(url, s, d)
}

func LogChild(parent Ghost, url, methodName string, d int64) Ghost {
	var spans []Span
	s := NewChildSpan(parent, methodName)
	s.Duration = d
	spans = append(spans, s)
	_, err := Send(url, spans)
	if err != nil {
		fmt.Println(err)
	}
	return s.ToGhost()

}

func main() {

	var spans []Span

	parent := NewSpan("101.test.com", "createPenguin")
	parent.Duration = 100000
	spans = append(spans, parent)
	time.Sleep(100 * time.Millisecond)

	child := NewChildSpan(parent.ToGhost(), "washPenguin")
	child.Duration = 200000
	spans = append(spans, child)

	body, err := Send("http://192.168.88.24:9411/api/v2/spans", spans)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(body))
	}

}
