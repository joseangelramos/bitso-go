package bitso

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joseangelramos/bitso-go/v3/common"
)

// Endpoints
const (
	baseAPIMainURL = "https://api.bitso.com"
	baseAPITestURL = "https://dev.bitso.com"
)

// UseTest switch all the API endpoints from production to the testnet
var UseTest = false

// Global enums
const (
	timestampKey = "timestamp"
	signatureKey = "signature"
)

func currentTimestamp() int64 {
	return FormatTimestamp(time.Now())
}

// FormatTimestamp formats a time into Unix timestamp in milliseconds
func FormatTimestamp(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// getAPIEndpoint return the base endpoint of the Rest API according the UseTestnet flag
func getAPIEndpoint() string {
	if UseTest {
		return baseAPITestURL
	}
	return baseAPIMainURL
}

// NewClient initialize an API client instance with API key and secret key.
// You should always call this function before using this SDK.
// Services will be created by the form client.NewXXXService().
func NewClient(apiKey, secretKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		SecretKey:  secretKey,
		BaseURL:    getAPIEndpoint(),
		UserAgent:  "Bitso-golang-api",
		HTTPClient: http.DefaultClient,
		Debug:      true,
		Logger:     log.New(os.Stderr, "Bitso-golang ", log.LstdFlags),
	}
}

type doFunc func(req *http.Request) (*http.Response, error)

// Client define API client
type Client struct {
	APIKey     string
	SecretKey  string
	BaseURL    string
	UserAgent  string
	HTTPClient *http.Client
	Debug      bool
	Logger     *log.Logger
	TimeOffset int64
	do         doFunc
}

func (c *Client) debug(format string, v ...interface{}) {
	if c.Debug {
		c.Logger.Printf(format, v...)
	}
}

func (c *Client) parseRequest(r *request, opts ...RequestOption) (err error) {
	// set request options from user
	for _, opt := range opts {
		opt(r)
	}
	err = r.validate()
	if err != nil {
		return err
	}

	fullURL := fmt.Sprintf("%s%s", c.BaseURL, r.endpoint)

	queryString := r.query.Encode()
	body := &bytes.Buffer{}
	bodyString := r.form.Encode()
	header := http.Header{}
	if r.header != nil {
		header = r.header.Clone()
	}
	if bodyString != "" {
		header.Set("Content-Type", "application/x-www-form-urlencoded")
		body = bytes.NewBufferString(bodyString)
	}

	if r.secType == secTypeSigned {

		nonce := currentTimestamp() + currentTimestamp()

		// Bitso API Secret on the concatenation of nonce + HTTP method + requestPath + JSON payload
		raw := fmt.Sprintf("%#v%s%s%s", nonce, r.method, r.endpoint, bodyString)

		mac := hmac.New(sha256.New, []byte(c.SecretKey))
		_, err = mac.Write([]byte(raw))
		if err != nil {
			return err
		}

		header.Set("Authorization", fmt.Sprintf("Bitso %s:%#v:%x", c.APIKey, nonce, (mac.Sum(nil))))

	}

	//if queryString == "" {
	//	queryString = v.Encode()
	//} else {
	//	queryString = fmt.Sprintf("%s&%s", queryString, v.Encode())
	//}

	if queryString != "" {
		fullURL = fmt.Sprintf("%s?%s", fullURL, queryString)
	}
	c.debug("full url: %s, body: %s", fullURL, bodyString)

	r.fullURL = fullURL
	r.header = header
	r.body = body
	return nil
}

func (c *Client) callAPI(ctx context.Context, r *request, opts ...RequestOption) (data []byte, err error) {
	err = c.parseRequest(r, opts...)
	if err != nil {
		return []byte{}, err
	}
	req, err := http.NewRequest(r.method, r.fullURL, r.body)
	if err != nil {
		return []byte{}, err
	}
	req = req.WithContext(ctx)
	req.Header = r.header
	c.debug("request: %#v", req)
	f := c.do
	if f == nil {
		f = c.HTTPClient.Do
	}
	res, err := f(req)
	if err != nil {
		return []byte{}, err
	}
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		cerr := res.Body.Close()
		// Only overwrite the returned error if the original error was nil and an
		// error occurred while closing the body.
		if err == nil && cerr != nil {
			err = cerr
		}
	}()
	c.debug("response: %#v", res)
	c.debug("response body: %s", string(data))
	c.debug("response status code: %d", res.StatusCode)

	if res.StatusCode >= http.StatusBadRequest {
		apiErr := new(common.APIError)
		e := json.Unmarshal(data, apiErr)
		if e != nil {
			c.debug("failed to unmarshal json: %s", e)
		}
		return nil, apiErr
	}
	return data, nil
}

// NewGetAccountService init getting account service
func (c *Client) NewGetAccountStatusService() *GetAccountStatusService {
	return &GetAccountStatusService{c: c}
}
