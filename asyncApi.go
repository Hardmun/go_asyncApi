package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	logFile     *os.File
	wg          sync.WaitGroup
	requestChan chan *requestStruct
	semaphore   chan struct{} // limit server requestPOOL
	doneRequest chan struct{} //signal to reduce threads
	requestPOOL = sync.Pool{New: func() interface{} { return new(requestStruct) }}
	absPath     string
)

type requestStruct struct {
	index      int
	httREQUEST *http.Request
	method     string
	errlist    []string
	result     []any
	url        string
	json       any
}

type jsonStruct struct {
	BaseURL  string            `json:"base_Url"`
	Url      string            `json:"url"`
	Ssl      any               `json:"ssl"`
	Login    string            `json:"login"`
	Password string            `json:"password"`
	Method   string            `json:"method"`
	ConnPool int               `json:"connPool"`
	Ord      string            `json:"ord"`
	Errlist  []string          `json:"errlist"`
	Headers  map[string]string `json:"headers"`
	Data     any               `json:"data"`
}

type errorDetails struct {
	Status       int    `json:"status"`
	StatusString string `json:"statusString"`
	Reason       string `json:"reason"`
	Url          string `json:"url"`
	Json         any    `json:"json"`
}

type errorStruct struct {
	Index int          `json:"index"`
	Error errorDetails `json:"error"`
}

type anyResponseSlice []map[string]any
type anyResponse map[string]any

func (ms *anyResponse) isMScoutError() error {
	if _, ok := (*ms)["ErrorType"]; !ok {
		return nil
	}
	if _, ok := (*ms)["ErrorItems"]; !ok {
		return nil
	}

	return errors.New((*ms)["ErrorItems"].([]interface{})[0].(map[string]interface{})[""+
		"ErrorMessage"].(string))
}

func (ms *anyResponse) isYandexError(statusCode *int) error {
	if errMsg, ok := (*ms)["error_message"].([]interface{}); *statusCode == 200 && ok {
		var errString string
		for _, errArray := range errMsg {
			if annotation, ok := errArray.(map[string]interface{})["annotation"].(string); ok {
				errString += annotation
			}
			if path, ok := errArray.(map[string]interface{})["paths"].([]interface{}); ok {
				for _, v := range path {
					if p, ok := v.(string); ok {
						errString += "\n(path: " + p + ")"
					}
				}
			}

		}
		return errors.New(errString)
	}

	if _, ok := (*ms)["detail"]; !ok {
		return nil
	}
	if _, ok := (*ms)["data"]; ok {
		return nil
	}

	if *statusCode == 404 {
		return errors.New("Result is empty")
	}

	return errors.New((*ms)["detail"].(string))
}

func (ms *anyResponse) isYandexOK(index *int) (interface{}, error) {
	if data, ok := (*ms)["data"]; ok {
		data.(map[string]any)["index"] = index
		for k, v := range *ms {
			if k != "data" {
				data.(map[string]any)[k] = v
			}
		}
		return data, nil
	}
	(*ms)["index"] = index
	return ms, nil
}

type resultStruct struct {
	Data []any `json:"data"`
}

func getErrorStructure(index, status *int, statusString, url *string, err *error,
	req interface{}, errlist *[]string) interface{} {

	newError := errorStruct{
		Error: errorDetails{
			Status:       *status,
			StatusString: *statusString,
			Reason:       (*err).Error(),
			Url:          *url,
			Json:         req,
		},
		Index: *index,
	}

	for _, e := range *errlist {
		if (strings.Contains(*statusString, e) || strings.Contains((*err).Error(), e)) && e != "" {
			loggErrorMessage(errors.New("Activating error:\n" + *statusString + "\n" + (*err).Error()))
			select {
			case doneRequest <- struct{}{}:
			default:
			}
			return nil
		}
	}
	return newError
}

func openFile(path string) (*os.File, error) {
	lFile, err := os.OpenFile(filepath.Join(absPath, path), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return lFile, nil
}

func systemError(errMsg error) {
	sysFile, errSys := openFile("sys.log")
	if errSys != nil {
		log.Fatal(errSys)
	}

	defer func(sysFile *os.File) {
		errClose := sysFile.Close()
		if errClose != nil {
			log.Fatal(errClose)
		}
	}(sysFile)

	errorLog := log.New(sysFile, "[error]", log.LstdFlags|log.Lshortfile)
	errorLog.Println(errMsg.Error())
}

func loggErrorMessage(err error) {
	errorLog := log.New(logFile, "[error]", log.LstdFlags|log.Lshortfile)
	errorLog.Println(err.Error())
}

func openJSON(path string) (*jsonStruct, error) {
	var (
		jsonFile *os.File
		err      error
		byteJSON []byte
	)

	jsonFile, err = os.Open(filepath.Join(absPath, path))
	if err != nil {
		return nil, errWrap(&err, "openJSON", "jsonFile, err = os.Open(path)"+"\n The pass is: "+
			filepath.Join(absPath, path))
	}

	byteJSON, err = io.ReadAll(jsonFile)
	if err != nil {
		return nil, errWrap(&err, "openJSON", "byteJSON, err = io.ReadAll(jsonFile)")
	}

	err = jsonFile.Close()
	if err != nil {
		return nil, errWrap(&err, "openJSON", "err = jsonFile.Close()")
	}

	if !json.Valid(byteJSON) {
		return nil, fmt.Errorf("invalid JSON string: %v", string(byteJSON))
	}

	var data jsonStruct
	err = json.Unmarshal(byteJSON, &data)
	if err != nil {
		return nil, errWrap(&err, "openJSON", "err = json.Unmarshal(byteJSON, &data)")
	}
	return &data, nil
}

func errWrap(err *error, fnc string, desc string) error {
	unpErr := *err
	return fmt.Errorf("%v\n (func: %v desc: %v)", unpErr.Error(), fnc, desc)
}

func httpREQUEST() {
	for dataFlow := range requestChan {
		semaphore <- struct{}{}

		go func(dataFlow *requestStruct) {
			var (
				client              *http.Client
				resp                *http.Response
				err                 error
				responseJSON        []byte
				defaultStatus       = 0
				defaultStrStatus    = ""
				responseStructSlice anyResponseSlice
				responseStruct      anyResponse
			)
			defer wg.Done()
			defer func() { <-semaphore }()

			client = &http.Client{}
			resp, err = client.Do(dataFlow.httREQUEST)
			if err != nil {
				if resp != nil {
					defaultStatus = resp.StatusCode
					defaultStrStatus = resp.Status
				}
				dataFlow.result[dataFlow.index] = getErrorStructure(&dataFlow.index, &defaultStatus,
					&defaultStrStatus, &dataFlow.url, &err, &dataFlow.json, &dataFlow.errlist)
				return
			}

			defer func(Body io.ReadCloser) {
				err = Body.Close()
				if err != nil {
					loggErrorMessage(err)
				}
			}(resp.Body)

			responseJSON, err = io.ReadAll(resp.Body)
			if err != nil {
				dataFlow.result[dataFlow.index] = getErrorStructure(&dataFlow.index, &defaultStatus,
					&defaultStrStatus, &dataFlow.url, &err, &dataFlow.json, &dataFlow.errlist)
				return
			}

			if errUnmSlice := json.Unmarshal(responseJSON, &responseStructSlice); errUnmSlice != nil {
				errUnm := json.Unmarshal(responseJSON, &responseStruct)
				if errUnm != nil {
					dataFlow.result[dataFlow.index] = getErrorStructure(&dataFlow.index, &resp.StatusCode,
						&resp.Status, &dataFlow.url, &errUnm, &dataFlow.json, &dataFlow.errlist)
					return
				}

				if err = responseStruct.isMScoutError(); err != nil {
					dataFlow.result[dataFlow.index] = getErrorStructure(&dataFlow.index, &resp.StatusCode,
						&resp.Status, &dataFlow.url, &err, &dataFlow.json, &dataFlow.errlist)
				} else if err = responseStruct.isYandexError(&resp.StatusCode); err != nil {
					dataFlow.result[dataFlow.index] = getErrorStructure(&dataFlow.index, &resp.StatusCode,
						&resp.Status, &dataFlow.url, &err, &dataFlow.json, &dataFlow.errlist)
				} else if yandexResp, isNil := responseStruct.isYandexOK(&dataFlow.index); isNil == nil {
					dataFlow.result[dataFlow.index] = yandexResp
				} else {
					responseStruct["index"] = dataFlow.index
					dataFlow.result[dataFlow.index] = responseStruct
				}
				return
			}

			switch ln := len(responseStructSlice); {
			case ln == 0:
				err = errors.New("Result is empty")
				dataFlow.result[dataFlow.index] = getErrorStructure(&dataFlow.index, &resp.StatusCode,
					&resp.Status, &dataFlow.url, &err, &dataFlow.json, &dataFlow.errlist)
			case ln == 1:
				responseStructSlice[0]["index"] = dataFlow.index
				dataFlow.result[dataFlow.index] = responseStructSlice[0]
			case ln > 1:
				err = errors.New(string(responseJSON))
				dataFlow.result[dataFlow.index] = getErrorStructure(&dataFlow.index, &resp.StatusCode,
					&resp.Status, &dataFlow.url, &err, &dataFlow.json, &dataFlow.errlist)
			}

		}(dataFlow)
	}
}

func callAsyncApi(uuid *string) error {
	var (
		errlist      []string
		data         *jsonStruct
		err          error
		basicAuth    bool
		allFilled    bool
		reqJSON      []byte
		responseJSON []byte
		reqAPI       *http.Request
		result       []any
	)

	data, err = openJSON(filepath.Join(*uuid, "data.json"))
	if err != nil {
		return errWrap(&err, "callAsyncApi", "data, err := openJSON(filepath.Join(*uuid, \"data.json\"))")
	}

	requests, ok := data.Data.([]interface{})
	if !ok {
		return fmt.Errorf("Cannot read request from JSON \n func: %v desc: %v",
			"callAsyncApi", "requests, ok := data.Data.([]interface{})")
	}

	resultLength := len(requests)
	//resultLength := 1 //TEST
	connPool := 50 //default
	if data.ConnPool != 0 {
		connPool = data.ConnPool
	}
	if len(data.Errlist) > 0 {
		errlist = data.Errlist
	}
	defaultCode := new(int)
	defaultStatus := new(string)

	//channels
	requestChan = make(chan *requestStruct)
	semaphore = make(chan struct{}, connPool)
	doneRequest = make(chan struct{}, 1)
	result = make([]any, resultLength)

	prefHTTP := "https://"
	if data.Ssl == false || data.Ssl == "false" {
		prefHTTP = "http://"
	}
	url := prefHTTP + data.BaseURL + data.Url
	login := data.Login
	password := data.Password
	headers := data.Headers
	method := strings.ToUpper(data.Method)

	//v := requests[0] //TEST
	go httpREQUEST()

labelMain:
	for {
		allFilled = true
	labelSlice:
		for k, v := range requests {
			//for k := 0; k < resultLength; k++ { //TEST
			if result[k] == nil {
				allFilled = false

				reqJSON, err = json.Marshal(&v)
				if err != nil {
					result[k] = getErrorStructure(&k, defaultCode, defaultStatus, &url, &err, &v, &errlist)
					continue
				}

				if method == "POST" {
					reqAPI, err = http.NewRequest(method, url, bytes.NewBuffer(reqJSON))
					if err != nil {
						result[k] = getErrorStructure(&k, defaultCode, defaultStatus, &url, &err, &v, &errlist)
						continue
					}
				} else if method == "GET" {
					reqAPI, err = http.NewRequest(method, url, nil)
					if err != nil {
						result[k] = getErrorStructure(&k, defaultCode, defaultStatus, &url, &err, &v, &errlist)
						continue
					}
					q := reqAPI.URL.Query()
					for parN, parV := range v.(map[string]interface{}) {
						q.Add(parN, parV.(string))
					}
					reqAPI.URL.RawQuery = q.Encode()

				} else {
					err = errors.New("Only two methods accepted :GET, POST.")
					result[k] = getErrorStructure(&k, defaultCode, defaultStatus, &url, &err, &v, &errlist)
					continue
				}

				basicAuth = true
				for key, value := range headers {
					reqAPI.Header.Set(key, value)
					if strings.ToLower(key) == "authorization" {
						basicAuth = false
					}
				}
				if basicAuth {
					reqAPI.SetBasicAuth(login, password)
				}

				newDataFlow := requestPOOL.Get().(*requestStruct)
				newDataFlow.index = k
				newDataFlow.httREQUEST = reqAPI
				newDataFlow.method = method
				newDataFlow.result = result
				newDataFlow.url = url
				newDataFlow.json = v
				newDataFlow.errlist = errlist

				wg.Add(1)
				requestChan <- newDataFlow
			}

			select {
			case <-doneRequest:
				{
					wg.Wait()
					//drain the channel
					for len(doneRequest) > 0 {
						<-doneRequest
					}

					if connPool == 1 {
						break labelMain
					} else {
						connPool = int(math.Floor(float64(connPool / 2)))
						loggErrorMessage(errors.New("Reduced threads to:" + strconv.Itoa(connPool)))
						semaphore = make(chan struct{}, connPool)
						break labelSlice
					}
				}
			default:
			}
		}
		wg.Wait()
		if allFilled {
			break labelMain
		}
	}

	close(requestChan)
	close(doneRequest)
	close(semaphore)

	resStruct := resultStruct{Data: result}
	responseJSON, err = json.Marshal(&resStruct)

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, responseJSON, "", "\t")

	err = os.WriteFile(filepath.Join(absPath, *uuid, "result.json"), prettyJSON.Bytes(), os.ModePerm)
	if err != nil {
		loggErrorMessage(err)
	}
	return nil
}

func clearLogs() {
	if err := os.Truncate(filepath.Join(absPath, "errors.log"), 0); err != nil {
		systemError(err)
	}
}

func clearTempFiles(uuid *string) {
	err := os.RemoveAll(filepath.Join(absPath, *uuid))
	if err != nil {
		loggErrorMessage(errWrap(&err, "clearTempFiles", "err := os.RemoveAll(*uuid)"))
	}
}

func main() {
	var err error

	//exePath, errExe := os.Executable()
	//if errExe != nil {
	//	log.Fatal(err)
	//}
	//absPath = filepath.Dir(exePath)

	logFile, err = openFile("errors.log")
	if err != nil {
		systemError(errWrap(&err, "main", "logFile, err = openFile(\"error.log\")"))
	}
	//Closing the logFile
	defer func(logFile *os.File) {
		err = logFile.Close()
		if err != nil {
			systemError(errWrap(&err, "main", "err = logFile.Close()"))
		}
	}(logFile)

	args := os.Args

	switch len(args) {
	case 2:
		arg := args[1]
		if arg == "-clearLogs" {
			clearLogs()
		} else {
			err = callAsyncApi(&arg)
			if err != nil {
				loggErrorMessage(errWrap(&err, "main", "err = callAsyncApi(&arg)"))
				fmt.Println(err.Error())
			}
		}
	case 3:
		if args[1] == "-clear" {
			clearTempFiles(&(args[2]))
		}
	default:
	}
	os.Exit(0)
}
