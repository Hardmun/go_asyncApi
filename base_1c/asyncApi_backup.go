package base_1c

//package main
//
//import (
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"io"
//	"log"
//	"net/http"
//	"os"
//	"path/filepath"
//	"sync"
//	"time"
//)
//
//var (
//	logFile *os.File
//	wg      sync.WaitGroup
//	//mtx     sync.Mutex
//)
//
//var glbParams struct {
//	url        string
//	login      string
//	password   string
//	reqHEADERS http.Header
//	//TODO: remove
//	writeFile bool
//}
//
//type jsonStruct struct {
//	BaseURL  string `json:"base_Url"`
//	Url      string `json:"url"`
//	Ssl      any    `json:"ssl"`
//	Login    string `json:"login"`
//	Password string `json:"password"`
//	Method   string `json:"method"`
//	ConnPool int    `json:"connPool"`
//	Ord      string `json:"ord"`
//	//Headers  struct {
//	//	ContentType string `json:"Content-type"`
//	//}
//	Headers map[string]string `json:"headers"`
//	Data    any               `json:"data"`
//}
//
//type resultStruct struct {
//	Data []any `json:"data"`
//}
//
//type errorDetails struct {
//	Status int    `json:"status"`
//	Reason string `json:"reason"`
//	Url    string `json:"url"`
//	Json   any    `json:"json"`
//}
//
//type errorStruct struct {
//	Error errorDetails `json:"error"`
//	Index int          `json:"index"`
//}
//
//type anyResponce []map[string]any
//
//func errorToStruct(index, status int, reason, url string, req interface{}) errorStruct {
//	newError := errorStruct{
//		Error: errorDetails{
//			Status: status,
//			Reason: reason,
//			Url:    url,
//			Json:   req,
//		},
//		Index: index,
//	}
//
//	return newError
//}
//
//func openFile(path string) (*os.File, error) {
//	lFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
//	if err != nil {
//		return nil, err
//	}
//	return lFile, nil
//}
//
//func systemError(errMsg error) {
//	sysFile, errSys := openFile("sys.log")
//	if errSys != nil {
//		log.Fatal(errSys)
//	}
//
//	defer func(sysFile *os.File) {
//		errClose := sysFile.Close()
//		if errClose != nil {
//			log.Fatal(errClose)
//		}
//	}(sysFile)
//
//	errorLog := log.New(sysFile, "[error]", log.LstdFlags|log.Lshortfile)
//	errorLog.Println(errMsg.Error())
//}
//
//func loggErrorMessage(err error) {
//	errorLog := log.New(logFile, "[error]", log.LstdFlags|log.Lshortfile)
//	errorLog.Println(err.Error())
//}
//
//func openJSON(path string) (*jsonStruct, error) {
//	var (
//		jsonFile *os.File
//		err      error
//		byteJSON []byte
//	)
//
//	jsonFile, err = os.Open(path)
//	if err != nil {
//		return nil, errWrap(&err, "openJSON", "jsonFile, err = os.Open(path)")
//	}
//
//	byteJSON, err = io.ReadAll(jsonFile)
//	if err != nil {
//		return nil, errWrap(&err, "openJSON", "byteJSON, err = io.ReadAll(jsonFile)")
//	}
//
//	err = jsonFile.Close()
//	if err != nil {
//		return nil, errWrap(&err, "openJSON", "err = jsonFile.Close()")
//	}
//
//	if !json.Valid(byteJSON) {
//		return nil, fmt.Errorf("invalid JSON string: %v", string(byteJSON))
//	}
//
//	var data jsonStruct
//	err = json.Unmarshal(byteJSON, &data)
//	if err != nil {
//		return nil, errWrap(&err, "openJSON", "err = json.Unmarshal(byteJSON, &data)")
//	}
//
//	return &data, nil
//}
//
//func errWrap(err *error, fnc string, desc string) error {
//	unpErr := *err
//	return fmt.Errorf("%v\n (func: %v desc: %v)", unpErr.Error(), fnc, desc)
//}
//
////backup
////func post(resultMap []any, k int, v any) {
////	var (
////		requestJSON  []byte
////		reqAPI       *http.Request
////		resp         *http.Response
////		responseJSON []byte
////		err          error
////		//defaultStatus = 0
////	)
////
////	defer wg.Done()
////
////	requestJSON, err = json.Marshal(&v)
////	if err != nil {
////		//resultMap[k] = errorToStruct(k, defaultStatus, err.Error(), glbParams.url, v)
////		return
////	}
////
////	reqAPI, err = http.NewRequest("POST", glbParams.url, bytes.NewBuffer(requestJSON))
////	if err != nil {
////		//resultMap[k] = errorToStruct(k, defaultStatus, err.Error(), glbParams.url, v)
////		return
////	}
////
////	reqAPI.Header = glbParams.reqHEADERS
////	//mtx.Lock()
////	reqAPI.SetBasicAuth(glbParams.login, glbParams.password)
////
////	//client := http.Client{}
////	//resp, err = client.Do(reqAPI)
////	//mtx.Unlock()
////	//if err != nil {
////	//	//if resp != nil {
////	//	//	defaultStatus = resp.StatusCode
////	//	//}
////	//	//resultMap[k] = errorToStruct(k, defaultStatus, err.Error(), glbParams.url, v)
////	//	return
////	//}
////	//defer resp.Body.Close()
////
////	//responseJSON, err = io.ReadAll(resp.Body)
////	//if err != nil {
////	//	//resultMap[k] = errorToStruct(k, resp.StatusCode, err.Error(), glbParams.url, v)
////	//	return
////	//}
////	//
////	//var responseStruct anyResponce
////	//err = json.Unmarshal(responseJSON, &responseStruct)
////	//if err != nil {
////	//	//resultMap[k] = errorToStruct(k, resp.StatusCode, err.Error(), glbParams.url, v)
////	//	return
////	//}
////
////	//resultMap[k] = responseStruct
////
////	_ = requestJSON
////	_ = reqAPI
////	_ = resp
////	_ = responseJSON
////	_ = err
////	//_ = client
////	//_ = resp1
////	//_ = err1
////}
//
//func post(resultMap []any, k int, v any, reqAPI *http.Request) {
//	var (
//		requestJSON []byte
//		//reqAPI       *http.Request
//		resp          *http.Response
//		responseJSON  []byte
//		err           error
//		defaultStatus = 0
//	)
//
//	defer wg.Done()
//
//	requestJSON, err = json.Marshal(&v)
//	if err != nil {
//		resultMap[k] = errorToStruct(k, defaultStatus, err.Error(), glbParams.url, v)
//		return
//	}
//
//	reqAPI.Body = io.NopCloser(bytes.NewBuffer(requestJSON))
//	//reqAPI, err = http.NewRequest("POST", glbParams.url, bytes.NewBuffer(requestJSON))
//	//if err != nil {
//	//	//resultMap[k] = errorToStruct(k, defaultStatus, err.Error(), glbParams.url, v)
//	//	return
//	//}
//	//
//	//reqAPI.Header = glbParams.reqHEADERS
//	////mtx.Lock()
//	//reqAPI.SetBasicAuth(glbParams.login, glbParams.password)
//
//	client := http.Client{}
//	resp, err = client.Do(reqAPI)
//	if err != nil {
//		if resp != nil {
//			defaultStatus = resp.StatusCode
//		}
//		resultMap[k] = errorToStruct(k, defaultStatus, err.Error(), glbParams.url, v)
//		return
//	}
//	defer resp.Body.Close()
//
//	responseJSON, err = io.ReadAll(resp.Body)
//	if err != nil {
//		resultMap[k] = errorToStruct(k, resp.StatusCode, err.Error(), glbParams.url, v)
//		return
//	}
//
//	var responseStruct anyResponce
//	err = json.Unmarshal(responseJSON, &responseStruct)
//	if err != nil {
//		if responseJSON != nil && glbParams.writeFile {
//			var responseStruct anyResponce
//			fmt.Println(string(responseJSON))
//			err = json.Unmarshal(responseJSON, &responseStruct)
//			os.WriteFile("somethig.json", responseJSON, os.ModePerm)
//			glbParams.writeFile = false
//		}
//
//		errDesc := err.Error()
//		if responseJSON != nil {
//			errDesc = errWrap(&err, "asyncApi", "err = json.Unmarshal(responseJSON, &responseStruct):"+
//				string(responseJSON)).Error()
//		}
//		resultMap[k] = errorToStruct(k, resp.StatusCode, errDesc, glbParams.url, v)
//		return
//	}
//
//	resultMap[k] = responseStruct
//
//	_ = requestJSON
//	_ = reqAPI
//	_ = resp
//	_ = responseJSON
//	_ = err
//	//_ = client
//	//_ = resp1
//	//_ = err1
//}
//
//func callAsyncApi(uuid *string) error {
//	var (
//		data         *jsonStruct
//		err          error
//		responseJSON []byte
//		reqAPI       *http.Request
//	)
//
//	tBegin := time.Now()
//	fmt.Printf("start: %v\n", tBegin)
//
//	data, err = openJSON(filepath.Join(*uuid, "data.json"))
//	if err != nil {
//		return errWrap(&err, "callAsyncApi", "data, err := openJSON(filepath.Join(*uuid, \"data.json\"))")
//	}
//
//	requests, ok := data.Data.([]interface{})
//	if !ok {
//		return fmt.Errorf("Cannot read request from JSON \n func: %v desc: %v",
//			"callAsyncApi", "requests, ok := data.Data.([]interface{})")
//	}
//
//	//resultLength := len(requests)
//	resultLength := 100
//	resultMap := make([]any, resultLength)
//	prefHTTP := "https://"
//	if data.Ssl == false || data.Ssl == "false" {
//		prefHTTP = "http://"
//	}
//	glbParams.url = prefHTTP + data.BaseURL + data.Url
//	glbParams.login = data.Login
//	glbParams.password = data.Password
//	//TODO: remove
//	glbParams.writeFile = true
//
//	var reqHEADERS = make(http.Header, len(data.Headers))
//	for k, v := range data.Headers {
//		reqHEADERS.Set(k, v)
//	}
//	glbParams.reqHEADERS = reqHEADERS
//
//	//TODO: testing
//	reqAPI, err = http.NewRequest("POST", glbParams.url, bytes.NewBuffer([]byte{}))
//	if err != nil {
//		//resultMap[k] = errorToStruct(k, defaultStatus, err.Error(), glbParams.url, v)
//		return err
//	}
//
//	reqAPI.Header = glbParams.reqHEADERS
//	reqAPI.SetBasicAuth(glbParams.login, glbParams.password)
//
//	v := requests[0]
//	wg.Add(resultLength)
//	//for k, v := range requests {
//	for k := 0; k < resultLength; k++ {
//		go post(resultMap, k, v, reqAPI)
//	}
//
//	wg.Wait()
//
//	resStruct := resultStruct{Data: resultMap}
//	responseJSON, err = json.Marshal(&resStruct)
//
//	var prettyJSON bytes.Buffer
//	err = json.Indent(&prettyJSON, responseJSON, "", "\t")
//
//	os.WriteFile(filepath.Join(*uuid, "result.json"), prettyJSON.Bytes(), os.ModePerm)
//
//	tEnd := time.Now()
//	fmt.Printf("End: %v\nDuration: %v", tEnd, tEnd.Sub(tBegin))
//
//	return nil
//}
//
//func main() {
//	var err error
//	logFile, err = openFile("error.log")
//	if err != nil {
//		systemError(errWrap(&err, "main", "logFile, err = openFile(\"error.log\")"))
//	}
//	//Closing the logFile
//	defer func(logFile *os.File) {
//		err = logFile.Close()
//		if err != nil {
//			systemError(errWrap(&err, "main", "err = logFile.Close()"))
//		}
//	}(logFile)
//
//	args := os.Args
//
//	switch len(args) {
//	case 2:
//		arg := args[1]
//
//		if arg == "-clearLogs" {
//			//TODO: Clear log with parameter
//		} else {
//			err = callAsyncApi(&arg)
//			if err != nil {
//				loggErrorMessage(errWrap(&err, "main", "err = callAsyncApi(&arg)"))
//				fmt.Println(err.Error())
//			}
//		}
//	case 3:
//		//TODO: Clear log with parameter - folder
//	default:
//		os.Exit(0)
//	}
//}
