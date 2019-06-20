package proxy

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gorilla/sessions"
)

// Config ...
type Config struct {
	ReverseProxy   *httputil.ReverseProxy
	Domain         string
	ClientID       string
	ClientSecret   string
	RedirectURI    string
	SessionSecret  []byte
	SessionTimeout time.Duration
}

// Auth0Proxy ...
type Auth0Proxy struct {
	ReverseProxy   *httputil.ReverseProxy
	Domain         string
	ClientID       string
	ClientSecret   string
	RedirectURI    string
	SessionTimeout time.Duration
	store          *sessions.CookieStore
	requests       map[string]*http.Request
}

// NewAuth0Proxy ...
func NewAuth0Proxy(c Config) *Auth0Proxy {
	return &Auth0Proxy{
		ReverseProxy:   c.ReverseProxy,
		Domain:         c.Domain,
		ClientID:       c.ClientID,
		ClientSecret:   c.ClientSecret,
		RedirectURI:    c.RedirectURI,
		SessionTimeout: c.SessionTimeout,
		store:          sessions.NewCookieStore(c.SessionSecret),
		requests:       map[string]*http.Request{},
	}
}

func (a *Auth0Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	session, err := a.store.Get(r, "auth0-proxy")
	if err != nil {
		fmt.Printf("[Debug] ServeHTTP Store get error. Error: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !session.IsNew {
		fmt.Printf("[Debug] ServeHTTP Handling Auth0 Callback\n")
		a.ReverseProxy.ServeHTTP(w, r)
		return
	}

	if code := r.URL.Query().Get("code"); code != "" {
		fmt.Printf("[Debug] ServeHTTP Code received: %s\n", code)

		a.handleAuth0Callback(w, r, code)
		return
	}

	a.handleAuth0Redirect(w, r)
}

func (a *Auth0Proxy) handleAuth0Redirect(w http.ResponseWriter, r *http.Request) {
	key := generateKey()
	a.requests[key] = r

	url := url.URL{
		Scheme:   "https",
		Host:     a.Domain,
		Path:     "/authorize",
		RawQuery: fmt.Sprintf("response_type=code&client_id=%s&redirect_uri=%s&state=%s", a.ClientID, a.RedirectURI, key),
	}

	fmt.Printf("[Debug] handleAuth0Redirect Redirecting to %s\n", url.String())

	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

func (a *Auth0Proxy) handleAuth0Callback(w http.ResponseWriter, r *http.Request, code string) {
	fmt.Printf("[Debug] handleAuth0Callback Handling Auth0 Callback\n")

	if err := a.validateCode(code); err != nil {
		fmt.Printf("[Debug] handleAuth0Callback Code not valid. Error: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := a.store.Get(r, "auth0-proxy")
	if err != nil {
		fmt.Printf("[Debug] handleAuth0Callback Store get error. Error: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Options.MaxAge = int(a.SessionTimeout.Seconds())
	session.Save(r, w)

	key := r.URL.Query().Get("state")
	if originalRequest := a.requests[key]; originalRequest != nil {
		fmt.Printf("[Debug] handleAuth0Callback Redirecting to %s\n", originalRequest.URL.String())
		delete(a.requests, key)
		http.Redirect(w, r, originalRequest.URL.String(), http.StatusSeeOther)
		return
	}

	a.ReverseProxy.ServeHTTP(w, r)
}

// CodeExchangeRequest ...
type CodeExchangeRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:'client_secret"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
}

func (a *Auth0Proxy) validateCode(code string) error {
	fmt.Printf("[Debug] validateCode\n")

	//client := rclient.NewRestClient(fmt.Sprintf("https://%s", a.Domain))

	//fmt.Printf("[Debug] validateCode NewRestClient to https://%s\n", a.Domain)

	req := CodeExchangeRequest{
		GrantType:    "authorization_code",
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		Code:         code,
		RedirectURI:  a.RedirectURI,
	}

	httpData := url.Values{}
	httpData.Set("grant_type", req.GrantType)
	httpData.Set("client_id", req.ClientID)
	httpData.Set("client_secret", req.ClientSecret)
	httpData.Set("code", req.Code)
	httpData.Set("redirect_uri", req.RedirectURI)

	fmt.Printf("[Debug] validateCode CodeExchangeRequest %+v\n", req)

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("https://%s/oauth/token", a.Domain), bytes.NewBufferString(httpData.Encode()))

	httpReq.Header.Set("User-Agent", "Auth0-GoProxy/1.0.0")
	httpReq.Header.Set("Accept", "*/*")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept-Encoding", "gzip,deflate")

	client := &http.Client{}
	resp, err := client.Do(httpReq)

	requestBodyResponse, _ := ioutil.ReadAll(resp.Request.Body)

	fmt.Printf("[Debug] validateCode Request Params %+v\n", resp.Request)
	fmt.Printf("[Debug] validateCode Request Params Body %+v\n", requestBodyResponse)

	if err != nil {
		fmt.Printf("[Debug] validateCode Post /oauth/token. Error: %s\n", err.Error())
		return err
	}
	defer resp.Body.Close()

	httpResponse, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("[Debug] validateCode Response %s\n", httpResponse)

	/*
		if err := client.Post("/oauth/token", req, nil); err != nil {
			fmt.Printf("[Debug] validateCode Post /oauth/token. Error: %s\n", err.Error())
			return err
		}
	*/

	return nil
}

func generateKey() string {
	salt := time.Now().Format(time.StampNano)
	return fmt.Sprintf("%x", md5.Sum([]byte(salt)))
}
