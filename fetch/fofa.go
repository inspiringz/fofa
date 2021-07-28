package fetch

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"fofa/logger"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	retryablehttp "github.com/projectdiscovery/retryablehttp-go"
)

var (
	errorUnauthorized = errors.New("401 Unauthorized, make sure email and apikey is correct.")
	errorForbidden    = errors.New("403 Forbidden, can't access the fofa service normally.")
	error503          = errors.New("503 Service Temporarily Unavailable")
	fields            = "host,ip,port,server,domain,title,country,province,city,icp"
	FetchResult       = make(map[string]([][]string))
)

var FetchResultT = struct {
	M map[string]([][]string)
	sync.Mutex
}{M: make(map[string]([][]string))}

type Fofa struct {
	email  string
	apiKey string
	size   string
	*retryablehttp.Client
}

func FofaRetryPolicy() func(ctx context.Context, resp *http.Response, err error) (bool, error) {
	return func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		// do not retry on context.Canceled or context.DeadlineExceeded
		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		if resp.StatusCode == 503 {
			return true, nil
		}

		return false, nil
	}
}

func RequestLogHook(req *http.Request, i int) {
	logger.Warn("Hook %v %v", i, req.URL)
}

func NewFofaClient(email, apiKey, size string) *Fofa {

	Client := retryablehttp.NewClient(retryablehttp.Options{
		// RetryWaitMin is the minimum time to wait for retry
		RetryWaitMin: 850 * time.Millisecond,
		// RetryWaitMax is the maximum time to wait for retry
		RetryWaitMax: 2400 * time.Millisecond,
		// Timeout is the maximum time to wait for the request
		Timeout: 15000 * time.Millisecond,
		// RetryMax is the maximum number of retries
		RetryMax: 6,
		// RespReadLimit is the maximum HTTP response size to read for
		// connection being reused.
		RespReadLimit: 4096,
		// Verbose specifies if debug messages should be printed
		Verbose: true,
		// KillIdleConn specifies if all keep-alive connections gets killed
		KillIdleConn: true,
	})
	Client.CheckRetry = FofaRetryPolicy()

	return &Fofa{
		email:  email,
		apiKey: apiKey,
		size:   size,
		Client: Client,
	}
}

func (ff *Fofa) Get(u string) ([]byte, error) {

	body, err := ff.Client.Get(u)
	if err != nil {
		return nil, err
	}
	defer body.Body.Close()

	if body.StatusCode == 503 {
		return nil, error503
	}

	content, err := ioutil.ReadAll(body.Body)

	if err != nil {
		return nil, err
	}
	return content, nil
}

func (ff *Fofa) Query(query string) (resultArray [][]string) {

	base64Query := base64.StdEncoding.EncodeToString([]byte(query))

	queryUrl := fmt.Sprintf("https://fofa.so/api/v1/search/all?email=%s&key=%s&qbase64=%s&size=%s&fields=%s",
		ff.email, ff.apiKey, base64Query, ff.size, fields)

	body, err := ff.Get(queryUrl)

	if err != nil {
		logger.Warn("<> " + query + " " + err.Error())
		return
	}

	results, _, _, err := jsonparser.Get(body, "results")

	if err != nil {
		logger.Warn("<> " + query + " " + err.Error())
		return
	}

	err = json.Unmarshal(results, &resultArray)
	if err != nil {
		logger.Warn("<> " + query + " " + err.Error())
		return
	}

	return
}

func (ff *Fofa) Auth() (valid bool, err error) {

	valid = false

	authUrl := fmt.Sprintf("https://fofa.so/api/v1/info/my?email=%s&key=%s", ff.email, ff.apiKey)

	body, err := ff.Get(authUrl)

	if err != nil {
		logger.Warn(err.Error())
		return
	}

	body_str := string(body)
	if strings.Contains(body_str, "fofa_server") {
		valid = true
	} else {
		if strings.Contains(body_str, "401") {
			err = errorUnauthorized
		} else if strings.Contains(body_str, "403") {
			err = errorForbidden
		}
		return
	}

	return
}

func (ff *Fofa) QueryAll(querys []string) {

	logger.Info(fmt.Sprintf("Totally %v rules waiting to be queried.", len(querys)))

	for i, q := range querys {
		logger.Info(fmt.Sprintf("Querying (%v, %v) now.", i, q))
		FetchResult[q] = ff.Query(q)
	}

	logger.Success(fmt.Sprintf("All (%v) Queries completed!", len(querys)))
}

func (ff *Fofa) QueryAllT(querys []string) {

	logger.Info(fmt.Sprintf("Totally %v rules waiting to be queried.", len(querys)))

	wg := sync.WaitGroup{}
	wg.Add(len(querys))

	for i, q := range querys {

		time.Sleep(75 * time.Millisecond)

		go func(pos int, query string) {
			defer wg.Done()
			logger.Info(fmt.Sprintf("-> %v, %v", pos, query))
			qresult := ff.Query(query)
			FetchResultT.Lock()
			FetchResultT.M[query] = qresult
			FetchResultT.Unlock()
			if len(qresult) == 0 {
				logger.Warn("<- %v, %v nothing found", pos, query)
			} else {
				logger.Success(fmt.Sprintf("<- %v, %v", pos, query))
			}
		}(i, q)
	}

	wg.Wait()
	logger.Success(fmt.Sprintf("All (%v) Queries completed!", len(querys)))
}
