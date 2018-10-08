package zipkin

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Span struct {
	TraceId     string              `json:"traceId"`
	Name        string              `json:"name"`
	ParentId    string              `json:"parentId"`
	ID          string              `json:"id"`
	Kind        string              `json:"kind"`
	Timestamp   int                 `json:"timestamp"`
	Duration    int                 `json:"duration"`
	Debug       bool                `json:"debug"`
	Shared      bool                `json:"shared"`
	Annotations []map[string]string `json:"annotations"`
	Tags        map[string]string   `json:"tags"`
}

type TinySpan struct {
	Name     string
	TraceId  string
	Duration int
}

func NewSpan(t TinySpan) *Span {
	s := Span{
		TraceId:   t.TraceId,
		Name:      t.Name,
		Timestamp: time.Now().Second(),
		Debug:     true,
		Shared:    true,
		Duration:  t.Duration,
		Kind:      "CLIENT",
	}
	u, _ := uuid.NewRandom()
	s.ID = u.String()

	s.Annotations = append(s.Annotations, make(map[string]string))
	//s.Annotations[0]["Key here"] = "value"

	s.Tags = make(map[string]string)
	//s.Tags["tag here"] = "value here"

	return &s

}

func main() {
	t := TinySpan{
		TraceId:  "100",
		Duration: 1000,
		Name:     "Service me"}
	s := NewSpan(t)

	b, _ := json.Marshal(s)
	fmt.Println(string(b))

}
