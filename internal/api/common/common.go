package common

import (
	"asyncApi/internal/input"
	"asyncApi/internal/logs"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"

	"fmt"
	"net/http"
	"strings"
	"sync"
)

var (
	semaphore chan struct{}
)

type requestParams struct {
	wg      *sync.WaitGroup
	method  string
	url     string
	headers map[string]string
	params  map[string]string
	errList []string
}

type response struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Result     any
}
type singleRequest struct {
	params requestParams
	body   []byte
	result *response
}

func internalErr(err error) response {
	return response{
		Status:     "Internal Server Error",
		StatusCode: 500,
		Result:     err.Error(),
	}
}

func doRequest(requestData *singleRequest) {
	defer requestData.params.wg.Done()

	var (
		apiResponse []byte
		resp        *http.Response
	)

	apiParams := requestData.params
	result := requestData.result

	request, err := http.NewRequest(apiParams.method, apiParams.url, bytes.NewBuffer(requestData.body))
	if err != nil {
		*result = internalErr(err)
		return
	}
	for key, value := range apiParams.headers {
		request.Header.Set(key, value)
	}
	p := request.URL.Query()
	for pName, pVal := range apiParams.params {
		p.Add(pName, pVal)
	}
	request.URL.RawQuery = p.Encode()

	client := &http.Client{}
	resp, err = client.Do(request)
	if err != nil {
		*result = internalErr(err)
		return
	}
	defer func(b io.ReadCloser) {
		err = b.Close()
		if err != nil {
			errLog, _ := logs.GetErrorLog()
			errLog.Write(err)
		}
	}(resp.Body)

	apiResponse, err = io.ReadAll(resp.Body)
	if err != nil {
		*result = response{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Result:     err.Error(),
		}
	}

	base64Data := base64.StdEncoding.EncodeToString(apiResponse)
	*result = response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Result:     base64Data,
	}
}

func CallAsyncApi(query *input.InpParams) error {
	var (
		allFilled bool
		err       error
		reqJSON   []byte
		requests  []any
		res       []response
	)

	switch query.Body.(type) {
	case map[string]any:
		requests = []any{query.Body}
	case []any:
		requests = query.Body.([]any)
	default:
		return fmt.Errorf("%s", "Cannot read the request from JSON")
	}

	resLen := len(requests)
	res = make([]response, resLen)

	connPool := 50 //default
	if query.ConnPool != 0 {
		connPool = query.ConnPool
	}
	semaphore = make(chan struct{}, connPool)

	p := "https://"
	if !query.Ssl {
		p = "http://"
	}

	reqParams := requestParams{
		wg:      &sync.WaitGroup{},
		method:  strings.ToUpper(query.Method),
		url:     p + query.Server + query.EndPoint,
		headers: query.Headers,
		params:  query.Params,
		errList: query.Errlist,
	}

labelMain:
	for {
		allFilled = true
		//labelSlice:
		for k, v := range requests {
			if res[k].StatusCode == 0 {
				allFilled = false

				reqJSON, err = json.Marshal(&v)
				if err != nil {
					res[k] = internalErr(err)
					continue
				}

				requestData := singleRequest{
					params: reqParams,
					body:   reqJSON,
					result: &res[k]}

				reqParams.wg.Add(1)
				//TODO: make goroutine
				go doRequest(&requestData)
			}
		}

		reqParams.wg.Wait()
		if allFilled {
			break labelMain
		}
	}

	return nil
}
