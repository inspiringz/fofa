package fetch

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"fofa/logger"
	"hash"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	retryablehttp "github.com/projectdiscovery/retryablehttp-go"
	"github.com/twmb/murmur3"
)

var (
	errorUnauthorized = errors.New("401 Unauthorized, make sure email and apikey is correct.")
	errorForbidden    = errors.New("403 Forbidden, can't access the fofa service normally.")
	error503          = errors.New("503 Service Temporarily Unavailable")
	error502          = errors.New("502 Bad Gateway")
	fields            = "host,ip,port,server,domain,title,country,province,city,icp,protocol"
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
	logger.Debug("--> REQ %v %v", i, req.URL)
}

func ResponseLogHook(resp *http.Response) {
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warn(err.Error())
	}
	logger.Debug("<-- RESP %v", resp.StatusCode, string(content))
}

func NewFofaClient(email, apiKey, size string) *Fofa {

	Client := retryablehttp.NewClient(retryablehttp.Options{
		// RetryWaitMin is the minimum time to wait for retry
		RetryWaitMin: 700 * time.Millisecond,
		// RetryWaitMax is the maximum time to wait for retry
		RetryWaitMax: 3500 * time.Millisecond,
		// Timeout is the maximum time to wait for the request
		Timeout: 14000 * time.Millisecond,
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
	//Client.RequestLogHook = RequestLogHook
	//Client.ResponseLogHook = ResponseLogHook

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
	} else if body.StatusCode == 502 {
		return nil, error502
	} else if body.StatusCode == 403 {
		return nil, errorForbidden
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
		logger.Warn("<< " + query + " " + err.Error())
		return
	}

	errmsg, err := jsonparser.GetString(body, "errmsg")

	if errmsg != "" {
		logger.Warn("<< " + query + " " + errors.New(errmsg).Error())
		return
	}

	results, _, _, err := jsonparser.Get(body, "results")

	if err != nil {
		logger.Warn("<< " + query + " " + err.Error())
		return
	}

	err = json.Unmarshal(results, &resultArray)
	if err != nil {
		logger.Warn("<< " + query + " " + err.Error())
		return
	}

	return
}

func (ff *Fofa) Auth() (valid bool, err error) {

	valid = false
	err = nil

	authUrl := fmt.Sprintf("https://fofa.so/api/v1/info/my?email=%s&key=%s", ff.email, ff.apiKey)

	body, err := ff.Get(authUrl)

	if err != nil {
		return
	}

	body_str := string(body)

	if strings.Contains(body_str, "fofa_server") {
		valid = true
		return
	} else {
		if strings.Contains(body_str, "401") {
			err = errorUnauthorized
		} else if strings.Contains(body_str, "403") {
			err = errorForbidden
		}
		return
	}
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

		time.Sleep(90 * time.Millisecond)

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

func (ff *Fofa) IconHash(url string) (iconHash string) {

	content, err := ff.Get(url)

	if err != nil {
		logger.Warn(err.Error())
		return
	}

	b64s := base64.StdEncoding.EncodeToString(content)

	var buffer bytes.Buffer

	for i := 0; i < len(b64s); i++ {
		ch := b64s[i]
		buffer.WriteByte(ch)
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}

	buffer.WriteByte('\n')

	var h32 hash.Hash32 = murmur3.New32()
	h32.Write(buffer.Bytes())
	return fmt.Sprintf("%d", int32(h32.Sum32()))
}

func GetIconHash(url string) {

	logger.Info(url)

	iconHash := NewFofaClient("", "", "").IconHash(url)

	logger.Success("icon_hash=\"%v\"", iconHash)
}

func (ff *Fofa) FofaTip(keyword string) (dropList []string) {

	tipUrl := "https://api.fofa.so/v1/search/tip?q=" + keyword

	content, err := ff.Get(tipUrl)

	if err != nil {
		logger.Warn(err.Error())
		return nil
	}

	regexp := regexp.MustCompile(`"name":"[^"]+`)
	results := regexp.FindAllStringSubmatch(string(content), -1)

	for _, v := range results {
		dropList = append(dropList, fmt.Sprintf("%v", v[0][8:]))
	}

	return
}

func GetFofaTip(keyword string) {

	logger.Info(fmt.Sprintf("Keyword: %v", keyword))

	dropList := NewFofaClient("", "", "").FofaTip(keyword)

	if dropList == nil {
		logger.Warn("Nothing found, try another keyword plz.")
		return
	}

	for _, v := range dropList {
		logger.Success(v)
	}
}
