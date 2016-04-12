package v2

import (
	"bytes"
	"crypto/tls"
	//	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"

	//	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrNotFound        = errors.New("Not found")
	defaultHTTPTimeout = 30 * time.Second
)

type RegistryClient struct {
	URL        *url.URL
	tlsConfig  *tls.Config
	httpClient *http.Client
	User       string
}

type Repo struct {
	Namespace  string
	Repository string
}

//type RegistryUser struct {
//	User     string
//	Password string
//}

//func NewRegistryUser(user, password string) *RegistryUser {

//	return &RegistryUser{
//		User:     user,
//		Password: password,
//	}

//}

func newHTTPClient(u *url.URL, tlsConfig *tls.Config, timeout time.Duration) *http.Client {
	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpTransport.Dial = func(proto, addr string) (net.Conn, error) {
		return net.DialTimeout(proto, addr, timeout)
	}
	return &http.Client{Transport: httpTransport}
}

func NewRegistryClient(registryUrl string, tlsConfig *tls.Config) (*RegistryClient, error) {
	u, err := url.Parse(registryUrl)
	if err != nil {
		return nil, err
	}
	httpClient := newHTTPClient(u, tlsConfig, defaultHTTPTimeout)
	return &RegistryClient{
		URL:        u,
		httpClient: httpClient,
		tlsConfig:  tlsConfig,
	}, nil
}

func (client *RegistryClient) doRequest(method string, path, user, passwd string, body []byte, headers map[string]string) ([]byte, http.Header, error) {
	b := bytes.NewBuffer(body)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client1 := &http.Client{Transport: tr}
	remta := strings.Split(client.URL.String(), "//")
	if len(remta) != 2 {
		return nil, nil, nil
	}
	remtan := remta[1]

	remtanurl := fmt.Sprintf("https://%s:%s@%s", user, passwd, remtan)
	req, err := http.NewRequest(method, remtanurl+"/v2"+path, b)

	if err != nil {
		return nil, nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	if headers != nil {
		for header, value := range headers {
			req.Header.Add(header, value)
		}
	}

	resp, err := client1.Do(req)
	if err != nil {
		if !strings.Contains(err.Error(), "connection refused") && client.tlsConfig == nil {
			return nil, nil, fmt.Errorf("%v. Are you trying to connect to a TLS-enabled daemon without TLS?", err)
		}
		return nil, nil, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode == 404 {
		return nil, nil, ErrNotFound
	}

	if resp.StatusCode >= 400 {
		return nil, nil, Error{StatusCode: resp.StatusCode, Status: resp.Status, msg: string(data)}
	}
	return data, resp.Header, nil
}

func (client *RegistryClient) Search(query, user, passwd string) ([]*Repository, error) {
	type repo struct {
		Repositories []string `json:"repositories"`
	}

	uri := fmt.Sprintf("/_catalog")
	data, _, err := client.doRequest("GET", uri, user, passwd, nil, nil)
	if err != nil {
		return nil, err
	}

	res := &repo{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	repos := []*Repository{}
	// simple filter for list
	for _, k := range res.Repositories {
		if strings.Index(k, query) == 0 {
			type tagList struct {
				Tags []string `json:"tags"`
			}

			uri := fmt.Sprintf("/%s/tags/list", k)
			data, _, err := client.doRequest("GET", uri, user, passwd, nil, nil)
			if err != nil {
				return nil, err
			}

			tl := &tagList{}
			if err := json.Unmarshal(data, &tl); err != nil {
				return nil, err
			}
			for _, t := range tl.Tags {
				// get the repository and append to the slice
				r, err := client.Repository(k, t, user, passwd)
				if err != nil {
					return nil, err
				}

				repos = append(repos, r)
			}
		}
	}

	return repos, nil
}

func (client *RegistryClient) DeleteRepository(repo string) error {
	//uri := fmt.Sprintf("/repositories/%s/%s/", r.Namespace, r.Repository)
	//if _, _, err := client.doRequest("DELETE", uri, nil, nil); err != nil {
	//	return err
	//}

	return nil
}

func (client *RegistryClient) DeleteTag(repo string, tag, user, passwd string) error {
	r, err := client.Repository(repo, tag, user, passwd)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("/%s/manifests/%s", repo, r.Digest)
	if _, _, err := client.doRequest("DELETE", uri, user, passwd, nil, nil); err != nil {
		return err
	}

	return nil
}

func (client *RegistryClient) Repository(name, tag, user, passwd string) (*Repository, error) {
	if tag == "" {
		tag = "latest"
	}

	uri := fmt.Sprintf("/%s/manifests/%s", name, tag)

	data, hdr, err := client.doRequest("GET", uri, user, passwd, nil, nil)
	if err != nil {
		return nil, err
	}
	repo := &Repository{}
	if err := json.Unmarshal(data, &repo); err != nil {
		return nil, err
	}

	repo.Digest = hdr.Get("Docker-Content-Digest")
	return repo, nil
}
