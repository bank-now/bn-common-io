package rest

import (
	"bytes"
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

func Post(url string, b []byte) (body []byte, err error) {
	response, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return
	}
	defer response.Body.Close()
	return ioutil.ReadAll(response.Body)

}
