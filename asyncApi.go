package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile *os.File
	wg      sync.WaitGroup
)

type connectionParams struct {
	url      string
	login    string
	password string
	headers  map[string]string
}

// limit server requests
var simaphore chan struct{}

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

type resultStruct struct {
	Data []any `json:"data"`
}

type errorDetails struct {
	Status int    `json:"status"`
	Reason string `json:"reason"`
	Url    string `json:"url"`
	Json   any    `json:"json"`
}

type errorStruct struct {
	Error errorDetails `json:"error"`
	Index int          `json:"index"`
}

type multResponse []map[string]any
type singleResponse map[string]any

func errorToStruct(index, status int, reason, url string, req interface{}) errorStruct {
	newError := errorStruct{
		Error: errorDetails{
			Status: status,
			Reason: reason,
			Url:    url,
			Json:   req,
		},
		Index: index,
	}

	return newError
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

func post(resultMap []any, k int, v any, params connectionParams) {
	var (
		requestJSON   []byte
		reqAPI        *http.Request
		resp          *http.Response
		responseJSON  []byte
		err           error
		defaultStatus = 0
	)

	defer wg.Done()

	//limit requests
	simaphore <- struct{}{}
	defer func() {
		<-simaphore
	}()

	requestJSON, err = json.Marshal(&v)
	if err != nil {
		resultMap[k] = errorToStruct(k, defaultStatus, err.Error(), params.url, v)
		return
	}

	reqAPI, err = http.NewRequest("POST", params.url, bytes.NewBuffer(requestJSON))
	if err != nil {
		resultMap[k] = errorToStruct(k, defaultStatus, err.Error(), params.url, v)
		return
	}

	for key, value := range params.headers {
		reqAPI.Header.Set(key, value)
	}
	reqAPI.SetBasicAuth(params.login, params.password)

	client := http.Client{}
	resp, err = client.Do(reqAPI)
	if err != nil {
		if resp != nil {
			defaultStatus = resp.StatusCode
		}
		resultMap[k] = errorToStruct(k, defaultStatus, err.Error(), params.url, v)
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
		resultMap[k] = errorToStruct(k, resp.StatusCode, err.Error(), params.url, v)
		return
	}

	var rawMessage json.RawMessage
	err = json.Unmarshal(responseJSON, &rawMessage)
	if err != nil {
		errDesc := err.Error()
		if responseJSON != nil {
			errDesc = errWrap(&err, "asyncApi", "err = json.Unmarshal(responseJSON, &responseStruct):"+
				string(responseJSON)).Error()
		}
		resultMap[k] = errorToStruct(k, resp.StatusCode, errDesc, params.url, v)
		return
	}

	//if _, ok := rawMessage[0]; !ok {
	//
	//}

	resultMap[k] = rawMessage[0] //responseStruct
}

func callAsyncApi(uuid *string) error {
	var (
		data         *jsonStruct
		err          error
		responseJSON []byte
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

	resultLength := len(requests)
	//resultLength := 500 //testing
	connPool := 50 //default
	if data.ConnPool != 0 {
		connPool = data.ConnPool
	}
	simaphore = make(chan struct{}, connPool)
	resultMap := make([]any, resultLength)
	prefHTTP := "https://"
	if data.Ssl == false || data.Ssl == "false" {
		prefHTTP = "http://"
	}

	connParams := connectionParams{
		url:      prefHTTP + data.BaseURL + data.Url,
		login:    data.Login,
		password: data.Password,
		headers:  data.Headers,
	}

	//v := requests[0] //testing
	wg.Add(resultLength)
	for k, v := range requests {
		//for k := 0; k < resultLength; k++ { //testing
		go post(resultMap, k, v, connParams)
	}

	wg.Wait()

	resStruct := resultStruct{Data: resultMap}
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
