package top

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/bitly/go-simplejson"
)

const (
	defaultSeparator = "\r\n-S-\r\n"
)

func (c *Client) doBatch(ctx context.Context, param Parameters, bodyStr string) ([]byte, error) {
	rbURL, err := url.Parse(c.opt.apiBatchURL)
	if err != nil {
		return nil, err
	}
	rbURL.RawQuery = param.getRequestData()
	req, err := http.NewRequestWithContext(ctx, "POST", rbURL.String(), strings.NewReader(bodyStr))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "text/plain;charset=UTF-8")
	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("%d:%s", response.StatusCode, response.Status)
		return nil, err
	}
	return ioutil.ReadAll(response.Body)
}

// BatchResult defines the batch result.
type BatchResult struct {
	*simplejson.Json
	Err error
}

// DoJsonBatch executes the taobao api in batch.
// https://open.taobao.com/doc.htm?docId=104350&docType=1
func (c *Client) DoJsonBatch(ctx context.Context, params ...Parameters) ([]*BatchResult, error) {
	p := Parameters{}
	p.setCommonParams(c.opt.appKey, c.opt.session)
	size := len(params)
	var reqBody bytes.Buffer
	for i, param := range params {
		reqBody.WriteString(param.getRequestData())
		if i != size-1 {
			reqBody.WriteString(defaultSeparator)
		}
	}
	bodyStr := reqBody.String()
	p.sign(c.opt.appSecret, bodyStr)

	bodyBytes, err := c.doBatch(ctx, p, bodyStr)
	if err != nil {
		return nil, err
	}
	return bytesToResultBatch(bodyBytes, size)
}

func bytesToResultBatch(bytes []byte, reqSize int) ([]*BatchResult, error) {
	ss := strings.Split(string(bytes), defaultSeparator)
	size := len(ss)
	res := make([]*BatchResult, size)
	for i, s := range ss {
		js, err := simplejson.NewJson([]byte(s))
		res[i] = &BatchResult{Json: js, Err: err}
		if err != nil {
			continue
		}
		res[i].Err = checkErrorResp(js)
	}
	if size != reqSize && size == 1 && res[0].Err != nil {
		// it's a common error
		return nil, res[0].Err
	}
	return res, nil
}
