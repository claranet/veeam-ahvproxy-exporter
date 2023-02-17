//
// AHVProxy-exporter
//
// Prometheus Exportewr for AHVProxy API
//
// Author: Martin Weber <martin.weber@de.clara.net>
// Company: Claranet GmbH
//

package veeam

import (
	//	"os"
	"crypto/tls"
	"net/http"
	"net/url"
	"strings"
	"fmt"
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type RequestParams struct {
	body, header string
	params       url.Values
}

type AHVProxy struct {
	url      string
	username string
	password string
	token		 string
}

func (g *AHVProxy) makeRequest(reqType string, action string) (*http.Response, error) {
	return g.makeRequestWithParams(reqType, action, RequestParams{})
}

func (g *AHVProxy) makeRequestWithParams(reqType string, action string, p RequestParams) (*http.Response, error) {
	_url := strings.Trim(g.url, "/")
	_url += "/"
	_url += strings.Trim(action, "/") + "/"

	log.Debugf("URL: %s", _url)
	log.Debugf("Payload: %v", p.body)
	
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	var netClient = http.Client{Transport: tr}

	body := p.body

	_url += "?" + p.params.Encode()

	req, err := http.NewRequest(reqType, _url, strings.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	
	// req.Header.Set("UserAgentInternal", "webfrontend/1.0")
	if len(g.token) > 0  {
		req.Header.Set("Content-Type", "text/json")
		req.Header.Set("Accept", "text/json")
	  req.Header.Set("Authorization", fmt.Sprintf(`Bearer %s`,g.token))
	}

	resp, err := netClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	if resp.StatusCode >= 400 {
		log.Fatal(resp.Status)
		return nil, nil
	}

	return resp, nil
}

func (g *AHVProxy) login() {
	_url := strings.Trim(g.url, "/")
	_url += "/api/oauth2/token"

	var payload = url.Values{}
	payload.Add("grantType", "password")
	payload.Add("userName", g.username)
	payload.Add("password", g.password)

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := http.Client{Transport: tr}

	resp, err := client.PostForm(_url, payload)
	if err != nil {
		log.Fatal(err)
		return
	}

	var data map[string]string
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&data)
	g.token = data["accessToken"]
}

func NewVeeamAhvProxy(url string, username string, password string) *AHVProxy {
	//	log.SetOutput(os.Stdout)
	//	log.SetPrefix("AHVProxy Logger")

	instance := &AHVProxy{
		url:      url,
		username: username,
		password: password,
	}
	instance.login()
	return instance
}
