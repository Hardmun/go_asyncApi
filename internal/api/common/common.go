package common

import (
	"asyncApi/internal/input"
	//"bytes"
	//"encoding/json"
	"errors"
	//"fmt"
	//"net/http"
	//"strings"
)

var (
	semaphore chan struct{}
)

type response struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Result     any
}

func internalErr(err error) response {
	return response{
		Status:     "Internal Server Error",
		StatusCode: 500,
		Result:     err.Error(),
	}
}

func CallAsyncApi(query *input.InpParams) error {
	//	var (
	//		allFilled bool
	//		err       error
	//		errlist   []string
	//		reqJSON   []byte
	//		request   *http.Request
	//		requests  []any
	//		result    []any
	//	)
	//
	//	switch query.Body.(type) {
	//	case map[string]any:
	//		requests = []any{query.Body}
	//	case []any:
	//		requests = query.Body.([]any)
	//	default:
	//		return fmt.Errorf("%s", "Cannot read the request from JSON")
	//	}
	//
	//	resLen := len(requests)
	//	result = make([]any, resLen)
	//
	//	connPool := 50 //default
	//	if query.ConnPool != 0 {
	//		connPool = query.ConnPool
	//	}
	//	semaphore = make(chan struct{}, connPool)
	//
	//	if len(query.Errlist) > 0 {
	//		errlist = query.Errlist
	//	}
	//
	//	p := "https://"
	//	if !query.Ssl {
	//		p = "http://"
	//	}
	//
	//	url := p + query.Server + query.EndPoint
	//
	//labelMain:
	//	for {
	//		allFilled = true
	//	labelSlice:
	//		for k, v := range requests {
	//			if result[k] == nil {
	//				allFilled = false
	//
	//				reqJSON, err = json.Marshal(&v)
	//				if err != nil {
	//					result[k] = internalErr(err)
	//					continue
	//				}
	//
	//				request, err = http.NewRequest(strings.ToUpper(query.Method), url, bytes.NewBuffer(reqJSON))
	//				if err != nil {
	//					result[k] = internalErr(err)
	//					continue
	//				}
	//
	//			}
	//		}
	//	}

	return errors.New("Has not done!")
}
