package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile     *os.File
	wg          sync.WaitGroup
	requestChan chan *requestStruct
	semaphore   chan struct{} // limit server requestPOOL
	doneRequest chan struct{} //signal to reduce threads
	requestPOOL = sync.Pool{New: func() interface{} { return new(requestStruct) }}
)

type requestStruct struct {
	index      int
	httREQUEST *http.Request
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

func (ms *anyResponse) isMScoutError() bool {
	if _, ok := (*ms)["ErrorType"]; !ok {
		return false
	}
	if _, ok := (*ms)["ErrorItems"]; !ok {
		return false
	}

	return true
}

type resultStruct struct {
	Data []any `json:"data"`
}

func getErrorStructure(index, status int, statusString, errDescription, url string, req interface{}) *errorStruct {
	newError := errorStruct{
		Error: errorDetails{
			Status:       status,
			StatusString: statusString,
			Reason:       errDescription,
			Url:          url,
			Json:         req,
		},
		Index: index,
	}

	select {
	case doneRequest <- struct{}{}:
	default:
	}

	_ = newError
	return nil
}

func openFile(path string) (*os.File, error) {
	lFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
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

	jsonFile, err = os.Open(path)
	if err != nil {
		return nil, errWrap(&err, "openJSON", "jsonFile, err = os.Open(path)")
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
		wg.Add(1)
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
				dataFlow.result[dataFlow.index] = getErrorStructure(dataFlow.index, defaultStatus, defaultStrStatus,
					err.Error(), dataFlow.url, dataFlow.json)
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
				dataFlow.result[dataFlow.index] = getErrorStructure(dataFlow.index, defaultStatus, defaultStrStatus,
					err.Error(), dataFlow.url, dataFlow.json)
				return
			}

			if errUnmSlice := json.Unmarshal(responseJSON, &responseStructSlice); errUnmSlice != nil {
				errUnm := json.Unmarshal(responseJSON, &responseStruct)
				if errUnm != nil {
					dataFlow.result[dataFlow.index] = getErrorStructure(dataFlow.index, resp.StatusCode, resp.Status,
						errUnm.Error(), dataFlow.url, dataFlow.json)
					return
				}

				if responseStruct.isMScoutError() {
					errMsg := responseStruct["ErrorItems"].([]interface{})[0].(map[string]interface{})[""+
						"ErrorMessage"].(string)
					dataFlow.result[dataFlow.index] = getErrorStructure(dataFlow.index, resp.StatusCode, resp.Status,
						errMsg, dataFlow.url, dataFlow.json)
				} else {
					responseStruct["index"] = dataFlow.index
					dataFlow.result[dataFlow.index] = responseStruct
				}
				return
			}

			switch ln := len(responseStructSlice); {
			case ln == 0:
				dataFlow.result[dataFlow.index] = getErrorStructure(dataFlow.index, resp.StatusCode, resp.Status,
					"Result is empty", dataFlow.url, dataFlow.json)
			case ln == 1:
				responseStructSlice[0]["index"] = dataFlow.index
				dataFlow.result[dataFlow.index] = responseStructSlice
			case ln > 1:
				dataFlow.result[dataFlow.index] = getErrorStructure(dataFlow.index, resp.StatusCode, resp.Status,
					string(responseJSON), dataFlow.url, dataFlow.json)
			}

		}(dataFlow)
	}
}

func callAsyncApi(uuid *string) error {
	var (
		data         *jsonStruct
		err          error
		allFilled    bool
		reqJSON      []byte
		responseJSON []byte
		reqAPI       *http.Request
		result       []any
	)

	tBegin := time.Now()
	fmt.Printf("start: %v\n", tBegin)

	data, err = openJSON(filepath.Join(*uuid, "data.json"))
	if err != nil {
		return errWrap(&err, "callAsyncApi", "data, err := openJSON(filepath.Join(*uuid, \"data.json\"))")
	}

	requests, ok := data.Data.([]interface{})
	if !ok {
		return fmt.Errorf("Cannot read request from JSON \n func: %v desc: %v",
			"callAsyncApi", "requests, ok := data.Data.([]interface{})")
	}

	//resultLength := len(requests)
	resultLength := 1800 //TEST
	connPool := 900      //default
	if data.ConnPool != 0 {
		connPool = data.ConnPool
	}

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

	v := requests[0] //TEST

	go httpREQUEST()

labelMain:
	for {
		allFilled = true
	labelSlice:
		//for i, v := range result {
		for k := 0; k < resultLength; k++ { //TEST
			if result[k] == nil {
				allFilled = false

				reqJSON, err = json.Marshal(&v)
				if err != nil {
					result[k] = getErrorStructure(k, 0, "", err.Error(), url, v)
					continue
				}

				reqAPI, err = http.NewRequest("POST", url, bytes.NewBuffer(reqJSON))
				if err != nil {
					result[k] = getErrorStructure(k, 0, "", err.Error(), url, v)
					continue
				}
				for key, value := range headers {
					reqAPI.Header.Set(key, value)
				}
				reqAPI.SetBasicAuth(login, password)

				newDataFlow := requestPOOL.Get().(*requestStruct)
				newDataFlow.index = k
				newDataFlow.httREQUEST = reqAPI
				newDataFlow.result = result
				newDataFlow.url = url
				newDataFlow.json = v
				requestChan <- newDataFlow
			}

			select {
			case <-doneRequest:
				{
					//drain the requestChan
					for len(requestChan) > 0 {
						<-requestChan
					}

					wg.Wait()

					//drain the channel
					for len(doneRequest) > 0 {
						<-doneRequest
					}

					if connPool == 1 {
						fmt.Printf("BREAK ALL connPool - %v\n", connPool)
						wg.Wait()
						break labelMain
					} else {
						connPool = int(math.Floor(float64(connPool / 2)))
						fmt.Printf("REDUSING connPool - %v\n", connPool)
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

	err = os.WriteFile(filepath.Join(*uuid, "result.json"), prettyJSON.Bytes(), os.ModePerm)
	if err != nil {
		loggErrorMessage(err)
	}

	tEnd := time.Now()
	fmt.Printf("End: %v\nDuration: %v", tEnd, tEnd.Sub(tBegin))

	return nil
}

func main() {
	var err error
	logFile, err = openFile("error.log")
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
			//TODO: Clear log with parameter
		} else {
			err = callAsyncApi(&arg)
			if err != nil {
				loggErrorMessage(errWrap(&err, "main", "err = callAsyncApi(&arg)"))
				fmt.Println(err.Error())
			}
		}
	case 3:
		//TODO: Clear log with parameter - folder
	default:
		os.Exit(0)
	}
}
