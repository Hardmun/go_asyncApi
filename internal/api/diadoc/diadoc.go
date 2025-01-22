package diadoc

import (
	cm "asyncApi/internal/api/common"
	"asyncApi/internal/crypt"
	in "asyncApi/internal/input"
	"asyncApi/utils"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	refreshOnce sync.Once
	semaphore   chan struct{}
	wg          sync.WaitGroup
)

type tokenParamsStruct struct {
	Url     string            `json:"url"`
	Client  string            `json:"client"`
	Secret  string            `json:"secret"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    map[string]string `json:"body"`
}

type reqStruct struct {
	tm          *tokenManager
	url         string
	method      string
	headers     tMap
	params      tMap
	result      *cm.Response
	downloadDir string
}
type tMap map[string]string

func (t *tMap) load(p map[string]interface{}) error {
	if k, ok := p["boxId"]; ok {
		v := k.(string)
		if v == "" {
			return fmt.Errorf("%s", "boxId is empty")
		}
		(*t)["boxId"] = v
	} else {
		return fmt.Errorf("%s", "boxId is missing")
	}

	if k, ok := p["messageId"]; ok {
		v := k.(string)
		if v == "" {
			return fmt.Errorf("%s", "messageId is empty")
		}
		(*t)["messageId"] = v
	} else {
		return fmt.Errorf("%s", "messageId is missing")
	}

	if k, ok := p["documentId"]; ok {
		v := k.(string)
		if v == "" {
			return fmt.Errorf("%s", "documentId is empty")
		}
		(*t)["documentId"] = v
	} else {
		return fmt.Errorf("%s", "documentId is missing")
	}

	return nil
}

type glbError struct {
	err any
	mu  sync.RWMutex
}

func (g *glbError) set(a any) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.err = a
}

func (g *glbError) get() any {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.err
}

func newGlobalError() *glbError {
	return &glbError{}
}

type tokenManager struct {
	tokens    *tMap
	urlParams *tokenParamsStruct
	mu        sync.RWMutex
	refreshCh chan struct{}
}

func (tm *tokenManager) initialize(tokenInfo string) error {
	rf, err := base64.StdEncoding.DecodeString(tokenInfo)
	if err != nil {
		return err
	}

	var tp tokenParamsStruct
	err = json.Unmarshal(rf, &tp)
	if err != nil {
		return err
	}

	tm.urlParams = &tp

	rf, err = os.ReadFile(filepath.Join(utils.GetDataPath(), "token", "auth.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if json.Valid(rf) {
		var (
			enc string
			tks tMap
		)
		err = json.Unmarshal(rf, &tks)
		if err != nil {
			return err
		}
		secret := tm.urlParams.Secret
		for k, v := range tks {
			enc, err = crypt.DecryptToken(v, secret)
			if err != nil {
				continue
			}
			tks[k] = enc
		}

		tm.tokens = &tks
	}

	if tm.get() == "" {
		err = tm.refresh()
		if err != nil {
			return err
		}
	}
	return nil
}

func (tm *tokenManager) set(token string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	(*tm.tokens)[tm.urlParams.Client] = token
}

func (tm *tokenManager) save() {
	tm.mu.RLock()
	t := make(tMap)
	secret := (*tm).urlParams.Secret
	var (
		b   []byte
		enc string
		err error
	)
	for k, v := range *tm.tokens {
		enc, err = crypt.EncryptToken(v, secret)
		if err != nil {
			t[k] = err.Error()
			continue
		}
		t[k] = enc
	}
	tm.mu.RUnlock()

	b, err = json.Marshal(t)
	if err != nil {
		return
	}

	var dirPath string
	dirPath, err = utils.DirPath(utils.GetDataPath(), "token")
	if err != nil {
		return
	}

	err = os.WriteFile(filepath.Join(dirPath, "auth.json"), b, os.ModePerm)
	if err != nil {
		return
	}
}

func (tm *tokenManager) refresh() error {
	var (
		req  *http.Request
		resp *http.Response
	)

	tm.mu.RLock()
	p := *tm.urlParams
	tm.mu.RUnlock()

	res, err := json.Marshal(p.Body)
	if err != nil {
		return err
	}

	req, err = http.NewRequest(p.Method, p.Url, bytes.NewBuffer(res))
	if err != nil {
		return err
	}

	if p.Client == "" {
		return fmt.Errorf("%s", "client key is missing")
	}

	req.Header.Set("Authorization", fmt.Sprintf("DiadocAuth ddauth_api_client_id=%s", p.Client))

	for k, v := range p.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return err
	}

	res, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return fmt.Errorf("%s, %s", resp.Status, string(res))
	}

	if resStr := string(res); resStr != "" {
		tm.set(resStr)
		tm.save()
	}

	return nil
}

func (tm *tokenManager) get() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if v, ok := (*tm.tokens)[tm.urlParams.Client]; ok {
		return v
	}

	return ""
}

func (tm *tokenManager) auth() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if v, ok := (*tm.tokens)[tm.urlParams.Client]; ok {
		return fmt.Sprintf("DiadocAuth ddauth_api_client_id=%s, ddauth_token=%s", tm.urlParams.Client, v)
	}

	return ""
}

func (tm *tokenManager) handle() {
	<-tm.refreshCh
}

func (tm *tokenManager) close() {
	close(tm.refreshCh)
}

func newTokenManager() *tokenManager {
	return &tokenManager{
		tokens:    &tMap{},
		refreshCh: make(chan struct{}),
	}
}

func doRequest(req reqStruct, isRetryAuth bool, isRetry bool) any {
	var (
		resp       *http.Response
		respReader []byte
	)

	request, err := http.NewRequest(req.method, req.url, nil)
	if err != nil {
		return err
	}
	for k, v := range req.headers {
		request.Header.Set(k, v)
	}
	request.Header.Set("Authorization", req.tm.auth())

	p := request.URL.Query()
	for k, v := range req.params {
		p.Add(k, v)
	}
	request.URL.RawQuery = p.Encode()

	client := &http.Client{}
	resp, err = client.Do(request)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("%s, %s", resp.Status, err.Error())
		}
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respReader, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	//refresh token on fail
	if resp.StatusCode == 401 {
		if isRetryAuth {
			req.result.Set(resp.Status, resp.StatusCode, string(respReader))
			return req.result
		} else {
			err = nil
			refreshOnce.Do(func() {
				err = req.tm.refresh()
				req.tm.close()
			})
			if err != nil {
				req.result.Set(resp.Status, resp.StatusCode, err.Error())
				return req.result
			}

			req.tm.handle()
			return doRequest(req, true, false)
		}
	}

	if resp.StatusCode != 200 {
		req.result.Set(resp.Status, resp.StatusCode, string(respReader))
	} else if rt := resp.Header.Get("Retry-After"); rt != "" && len(respReader) == 0 {
		if isRetry {
			req.result.Set(resp.Status, resp.StatusCode, string(respReader))
		} else {
			var rtn int
			rtn, err = strconv.Atoi(rt)
			if err != nil {
				req.result.Set(resp.Status, resp.StatusCode, fmt.Sprintf("Retry-After: %s", err.Error()))
			}
			time.Sleep(time.Duration(rtn+2) * time.Second)

			return doRequest(req, false, true)
		}
	} else {
		var path string
		path, err = utils.DirPath(req.downloadDir, req.params["boxId"],
			req.params["messageId"], req.params["documentId"])
		if err != nil {
			req.result.Set(resp.Status, resp.StatusCode, err.Error())
			return nil
		}

		err = os.WriteFile(filepath.Join(path, req.params["documentId"]+".pdf"), respReader, os.ModePerm)
		if err != nil {
			req.result.Set(resp.Status, resp.StatusCode, err.Error())
			return nil
		}

		req.result.Set(resp.Status, resp.StatusCode, "success")
	}

	return nil
}

func do(req reqStruct, glbErrChan chan any) {
	defer func() {
		<-semaphore
		wg.Done()
	}()

	err := doRequest(req, false, false)
	if err != nil {
		glbErrChan <- err
	}
}

func ufJson(tokenInfo string) any {
	var (
		isGlbErr int32
		requests []any
		res      []cm.Response
	)

	gError := newGlobalError()
	tm := newTokenManager()
	err := tm.initialize(tokenInfo)
	if err != nil {
		return err
	}

	params := in.InpParams
	commonParams := cm.GetCommonParams(params)

	if commonParams.Method != http.MethodGet {
		return fmt.Errorf("%s", "Only GET method is allowed")
	}

	if commonParams.DownloadDir == "" {
		return fmt.Errorf("%s", "DownloadDir must be specified")
	} else if _, err = os.Stat(commonParams.DownloadDir); err != nil {
		return err
	}

	switch params.Body.(type) {
	case map[string]any:
		requests = []any{params.Body}
	case string:
		requests = []any{struct{}{}}
	case []any:
		requests = params.Body.([]any)
	default:
		return fmt.Errorf("%s", "Cannot read the request from JSON")
	}

	semaphore = make(chan struct{}, in.InpParams.ConnPool)
	glbErrChan := make(chan any)
	go func() {
		for ge := range glbErrChan {
			if ge != nil {
				gError.set(ge)
				atomic.StoreInt32(&isGlbErr, 1)
			}
		}
	}()

	res = make([]cm.Response, len(requests))

	for k, v := range requests {
		semaphore <- struct{}{}

		if atomic.LoadInt32(&isGlbErr) == 1 {
			break
		}

		mp, ok := v.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s", "Wrong json structure. Must be the json {'key': 'value'}\n"+
				"key - boxId, messageId and documentId")
		}

		res[k].Index = k

		getParams := tMap{}
		err = getParams.load(mp)
		if err != nil {
			res[k].InternalErr(err)
			continue
		}

		req := reqStruct{
			tm:          tm,
			url:         commonParams.Url,
			method:      commonParams.Method,
			headers:     commonParams.Headers,
			params:      getParams,
			result:      &res[k],
			downloadDir: commonParams.DownloadDir,
		}

		wg.Add(1)
		go do(req, glbErrChan)
	}
	wg.Wait()
	close(glbErrChan)
	close(semaphore)

	if gErr := gError.get(); gErr != nil {
		return gErr
	}

	return res
}

func UploadFiles(tokenInfo string) error {
	res := ufJson(tokenInfo)

	switch r := res.(type) {
	case error:
		rsp := cm.Response{}
		rsp.InternalErr(r)
		res = rsp
	}

	err := cm.SaveResultFile(res, filepath.Join(in.WorkDir, "result.json"))
	if err != nil {
		return err
	}

	return nil
}
