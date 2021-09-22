package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/shiena/ansicolor"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	_url "net/url"
	"os"
	"strings"
)

var (
	paramsMap  = make(map[string]string)
	headersMap = map[string]string{
		"User-Agent":   "CloudApp/8.9.1 (com.bitqiu.pan; build:99; iOS 14.7.0) Alamofire/4.7.0",
		"Content-Type": "application/x-www-form-urlencoded",
	}
	apiHome = "https://pan-api.bitqiu.com"
)

type Resp struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Data    RespData `json:"data"`
}
type RespData struct {
	CurrentPage int        `json:"currentPage"`
	PageSize    int        `json:"pageSize"`
	Data        []Resource `json:"data"`
}

// Resource 资源信息
type Resource struct {
	ResourceId      string `json:"resourceId"`
	ResourceType    int    `json:"resourceType"`
	Name            string `json:"name"`
	Size            int    `json:"size"`
	DirLevel        int    `json:"dirLevel"`
	DirType         int    `json:"dirType"`
	CreateUid       int    `json:"createUid"`
	CreateTime      string `json:"createTime"`
	SnapTime        string `json:"snapTime"`
	ViewTime        string `json:"viewTime"`
	ViewOffsetMills string `json:"viewOffsetMills"`
	ResourceUid     string `json:"resourceUid"`
	Type            string `json:"type"`
}

func init() {
	// 改变默认的 Usage，flag包中的Usage 其实是一个函数类型。这里是覆盖默认函数实现，具体见后面Usage部分的分析
	flag.Usage = usage
	//InitLog 初始化日志
	log.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		ShowFullLevel:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
		// FieldsOrder: []string{"component", "category"},
	})
	// then wrap the log output with it
	log.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
	log.SetLevel(log.DebugLevel)

	// 初始化配置
	viper.SetConfigName("conf")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("read config failed: %v", err)
	}

	// 设置 headers 键值
	paramsMap["access_token"] = viper.GetString("validation.access_token")
	paramsMap["app_id"] = viper.GetString("validation.app_id")
	paramsMap["open_id"] = viper.GetString("validation.open_id")
	paramsMap["platform"] = viper.GetString("validation.platform")
	paramsMap["user_id"] = viper.GetString("validation.user_id")
}
func usage() {
	fmt.Fprintf(os.Stderr, `tiler version: tiler/v0.1.0
Usage: tiler [-h] [-c filename]
`)
	flag.PrintDefaults()
}

func HttpGet(url string, params map[string]string, headers map[string]string) (*http.Response, error) {
	//new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return nil, errors.New("new request is fail ")
	}
	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	//add headers
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	//http client
	client := &http.Client{}
	log.Printf("Go GET URL : %s \n", req.URL.String())
	return client.Do(req)
}
func HttpPost(url string, body map[string]string, params map[string]string, headers map[string]string) (*http.Response, error) {
	//add post body
	//var bodyJson []byte
	var req *http.Request
	var data = _url.Values{}
	if body != nil {
		for key, val := range body {
			data.Add(key, val)
		}
	}
	_body := strings.NewReader(data.Encode())
	req, err := http.NewRequest("POST", url, _body)
	if err != nil {
		log.Println(err)
		return nil, errors.New("new request is fail: %v \n")
	}

	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	//add headers
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	//requestDump, err := httputil.DumpRequest(req, true)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(string(requestDump))
	//http client
	client := &http.Client{}
	log.Printf("Go POST URL : %s \n", req.URL.String())
	return client.Do(req)
}

// 获取目录下的文件ids 最大二级
func getDirFileIds() []string {
	var fileIds []string
	paramsMap["parent_id"] = viper.GetString("file.dir_ids")
	paramsMap["desc"] = "1"
	paramsMap["limit"] = "1000"
	paramsMap["model"] = "1"
	paramsMap["order"] = "updateTime"
	paramsMap["page"] = "1"
	url := apiHome + "/fs/dir/resources/v2"
	resp, err := HttpPost(url, paramsMap, nil, headersMap)
	if err != nil {
		log.Fatal(err)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", content)
	var _resp Resp
	err = json.Unmarshal(content, &_resp)
	if err != nil {
		log.Fatal("format err:%s\n", err.Error())

	}
	fmt.Println(_resp)
	for _, val := range _resp.Data.Data {
		if val.Size > 200 {
			fileIds = append(fileIds, val.ResourceId)
		} else {
			// 进入下一级再获取文件资源id
			paramsMap["parent_id"] = val.ResourceId
			resp, err := HttpPost(url, paramsMap, nil, headersMap)
			if err != nil {
				log.Fatal(err)
			}
			content, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s\n", content)
			var _resp Resp
			err = json.Unmarshal(content, &_resp)
			for _, _val := range _resp.Data.Data {
				fmt.Println(_val.Size)
				if _val.Size > 200 {

					fileIds = append(fileIds, val.ResourceId)
				}
			}
		}
	}
	return fileIds
}

func main() {
	ids := getDirFileIds()

}
