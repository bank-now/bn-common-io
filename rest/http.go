package rest

import (
	"io/ioutil"
	"net/http"
)

func Get(url string) (body []byte, err error) {
	response, err := http.Get(url)
	if err != nil {
		return
	}
	defer response.Body.Close()
	return ioutil.ReadAll(response.Body)
}
