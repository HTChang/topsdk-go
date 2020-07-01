package top

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	simplejson "github.com/bitly/go-simplejson"
)

const (
	// api endpoint of open taobao platform.
	apiURL = "https://eco.taobao.com/router/rest"
	// batch api endpoint of open taobao platform.
	apiBatchURL = "https://eco.taobao.com/router/batch"
)

var (
	defaultOptions = options{
		apiURL:      apiURL,
		apiBatchURL: apiBatchURL,
		apiTimeout:  10 * time.Second,
	}
)

type options struct {
	appKey      string
	appSecret   string
	apiURL      string
	apiBatchURL string
	session     string
	apiTimeout  time.Duration
}

// ClientOption for building up `Client`.
type ClientOption func(opts *options)

// WithApiURL returns a `Option` that sets the api URL.
func WithApiURL(url string) ClientOption {
	return func(opts *options) {
		opts.apiURL = url
	}
}

// WithApiBatchURL returns a `Option` that sets the batch api URL.
func WithApiBatchURL(batchURL string) ClientOption {
	return func(opts *options) {
		opts.apiBatchURL = batchURL
	}
}

// WithSession returns a `Option` that set the session key.
// https://open.taobao.com/doc.htm?docId=102635&docType=1
func WithSession(session string) ClientOption {
	return func(opts *options) {
		opts.session = session
	}
}

// NewClient creates the `Client` according to the given options.
func NewClient(appKey, appSecret string, opts ...ClientOption) (*Client, error) {
	if appKey == "" {
		return nil, errors.New("app key cannot be empty")
	} else if appSecret == "" {
		return nil, errors.New("app secret cannot be empty")
	}

	opt := defaultOptions
	opt.appKey = appKey
	opt.appSecret = appSecret
	for _, o := range opts {
		o(&opt)
	}

	return &Client{
		opt: opt,
		httpClient: &http.Client{
			Timeout: opt.apiTimeout,
		},
	}, nil
}

// Client defines the open taobap client.
type Client struct {
	opt        options
	httpClient *http.Client
}

// Parameters defines the key value pair of parameters.
type Parameters map[string]interface{}

func (c *Client) do(ctx context.Context, param Parameters) (bytes []byte, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, "POST", c.opt.apiURL, strings.NewReader(param.getRequestData()))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	var response *http.Response
	response, err = c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("%d:%s", response.StatusCode, response.Status)
		return
	}
	bytes, err = ioutil.ReadAll(response.Body)
	return
}

// DoJson executes the taobao api.
// https://open.taobao.com/doc.htm?docId=101617&docType=1
func (c *Client) DoJson(ctx context.Context, method string, param Parameters) (res *simplejson.Json, err error) {
	param["method"] = method
	param.setCommonParams(c.opt.appKey, c.opt.session)
	param.sign(c.opt.appSecret, "" /* body */)

	var bodyBytes []byte
	bodyBytes, err = c.do(ctx, param)
	if err != nil {
		return
	}
	return bytesToResult(bodyBytes)
}

func bytesToResult(bytes []byte) (res *simplejson.Json, err error) {
	res, err = simplejson.NewJson(bytes)
	if err != nil {
		return
	}
	if err = checkErrorResp(res); err != nil {
		res = nil
	}
	return
}

func checkErrorResp(errResp *simplejson.Json) error {
	if resp, ok := errResp.CheckGet("error_response"); ok {
		bs, err := resp.Encode()
		if err != nil {
			return err
		}
		errResp := new(ErrorResponse)
		if err := json.Unmarshal(bs, errResp); err != nil {
			return err
		}
		return errResp
	}
	return nil
}

// setCommonParams sets the required parameters.
// https://open.taobao.com/doc.htm?docId=101617&docType=1
func (ps Parameters) setCommonParams(appKey, session string) {
	ps["app_key"] = appKey
	ps["timestamp"] = strconv.FormatInt(time.Now().In(cst).Unix(), 10)
	ps["format"] = "json"
	ps["v"] = "2.0"
	ps["sign_method"] = "md5"
	if session != "" {
		ps["session"] = session
	}

}

// sign gets the signature of params with optional body.
// https://open.taobao.com/doc.htm?docId=101617&docType=1
func (ps Parameters) sign(appSecret, body string) {
	keys := make([]string, 0, len(ps))
	for k := range ps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	query := bytes.NewBufferString(appSecret)
	for _, k := range keys {
		query.WriteString(k)
		query.WriteString(interfaceToString(ps[k]))
	}
	if body != "" {
		query.WriteString(body)
	}
	query.WriteString(appSecret)
	h := md5.New()
	io.Copy(h, query)
	ps["sign"] = strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func (ps Parameters) getRequestData() string {
	args := url.Values{}
	for key, val := range ps {
		args.Set(key, interfaceToString(val))
	}
	return args.Encode()
}
