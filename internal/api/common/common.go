package common

import (
	in "asyncApi/internal/input"
	"asyncApi/internal/logs"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	semaphore chan struct{}
)

type RequestParams struct {
	Wg          *sync.WaitGroup
	Method      string
	Url         string
	Headers     map[string]string
	Params      map[string]string
	ErrList     []string
	DownloadDir string
}

type Response struct {
	Index      int    `json:"index"`
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Result     any    `json:"result"`
}

func (r *Response) Set(status string, statusCode int, result string) {
	r.Status = status
	r.StatusCode = statusCode
	r.Result = result
}

func (r *Response) InternalErr(err error) {
	r.Set("Internal Server Error", 500, err.Error())
}

type SingleRequest struct {
	Params RequestParams
	Body   []byte
	Result *Response
}

func GetCommonParams(query *in.InpParamsStruct) RequestParams {
	if _, ok := query.Headers["Authorization"]; !ok && query.Login != "" {
		query.Headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(
			fmt.Sprintf("%s:%s", query.Login, query.Password)))
	}

	p := "https://"
	if !query.Ssl {
		p = "http://"
	}

	reqParams := RequestParams{
		Wg:          &sync.WaitGroup{},
		Method:      strings.ToUpper(query.Method),
		Url:         p + query.Server + query.EndPoint,
		Headers:     query.Headers,
		Params:      query.GetParams,
		ErrList:     query.Errlist,
		DownloadDir: query.DownloadDir,
	}

	return reqParams
}

func GetRequest(requestData *SingleRequest) (*http.Response, error) {

	var resp *http.Response

	apiParams := requestData.Params
	request, err := http.NewRequest(apiParams.Method, apiParams.Url, bytes.NewBuffer(requestData.Body))
	if err != nil {
		return nil, err
	}

	for key, value := range apiParams.Headers {
		request.Header.Set(key, value)
	}
	p := request.URL.Query()
	for pName, pVal := range apiParams.Params {
		p.Add(pName, pVal)
	}
	request.URL.RawQuery = p.Encode()

	client := &http.Client{}
	resp, err = client.Do(request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func doRequest(requestData *SingleRequest) {
	defer func() {
		<-semaphore
		requestData.Params.Wg.Done()
	}()

	var apiResponse []byte

	result := requestData.Result
	resp, err := GetRequest(requestData)
	if err != nil {
		result.InternalErr(err)
		//*result = InternalErr(err)
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
		result.InternalErr(err)
		//*result = InternalErr(err)
	}

	*result = Response{
		Index:      requestData.Result.Index,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Result:     base64.StdEncoding.EncodeToString(apiResponse),
	}
}

func SaveResultFile(res any, path string) error {
	respFile, err := json.Marshal(res)
	if err != nil {
		return err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, respFile, "", "\t")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, prettyJSON.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func CallAsyncApi(query *in.InpParamsStruct) error {
	var (
		allFilled bool
		err       error
		reqJSON   []byte
		requests  []any
		res       []Response
	)
	errLog, _ := logs.GetErrorLog()

	switch query.Body.(type) {
	case map[string]any:
		requests = []any{query.Body}
	case string:
		requests = []any{struct{}{}}
	case []any:
		requests = query.Body.([]any)
	default:
		return fmt.Errorf("%s", "Cannot read the request from JSON")
	}

	resLen := len(requests)
	res = make([]Response, resLen)

	connPool := 50 //default
	if query.ConnPool != 0 {
		connPool = query.ConnPool
	}
	semaphore = make(chan struct{}, connPool)
	reqParams := GetCommonParams(query)

labelMain:
	for {
		allFilled = true
		for k, v := range requests {
			if res[k].StatusCode == 0 {
				allFilled = false
				res[k].Index = k

				reqJSON, err = json.Marshal(&v)
				if err != nil {
					res[k].InternalErr(err)
					//res[k] = InternalErr(err)
					continue
				}

				requestData := SingleRequest{
					Params: reqParams,
					Body:   reqJSON,
					Result: &res[k]}

				semaphore <- struct{}{}
				reqParams.Wg.Add(1)
				//TODO: goroutine
				doRequest(&requestData)
			}
		}

		reqParams.Wg.Wait()
		if allFilled {
			break labelMain
		}
	}

	err = SaveResultFile(res, filepath.Join(in.WorkDir, "result.json"))
	if err != nil {
		errLog.Fatal(errLog)
	}

	return nil
}
