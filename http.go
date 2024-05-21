package mfesdk

import (
	"fmt"
	"time"

	"log"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/gconv"
)

type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func DoPostReq(urlStr string, reqBody []byte, cfg *MFECONF) (*Response, error) {
	ctx := gctx.New()
	client := g.Client()
	if cfg.IsProd {
		urlStr = BASE_API_URL + urlStr
	} else {
		urlStr = BASE_TEST_API_URL + urlStr
	}
	Sign, _ := signature(reqBody, cfg)
	timd := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("X-Security", "CFCA")
	client.SetHeader("X-AGENCY", cfg.AgencyNo)
	client.SetHeader("X-Sign", Sign)
	client.SetHeader("X-Time", timd)
	client.SetHeader("X-Trace", fmt.Sprint(time.Now().UnixNano()))
	req := gjson.MustEncode(g.Map{"param": string(reqBody)})
	res, err := client.Post(ctx, urlStr, req)
	if err != nil {
		g.Log().Error(ctx, err)
		return nil, err
	}
	defer res.Close()
	respSign := res.Header.Get("X-Sign")

	body := res.ReadAll()
	response := &Response{}
	err = gconv.Scan(body, response)
	if err != nil {
		g.Log().Error(ctx, err)
		return nil, err
	}
	if response.Code != "000000" {
		log.Println("请求返回失败 CODE:", response.Code, "请求返回失败 ERROR:", response.Message, "请求返回失败 DATA:", response.Data)
		return nil, fmt.Errorf(response.Message)
	}
	if ok, err := verify([]byte(response.Data), []byte(respSign), cfg); !ok {
		// 返回中文错误提示
		return nil, fmt.Errorf("响应签名验证失败:%s", err)
	}
	return response, nil
}

/*
urlStr 请求路径
reqBody 请求参数
*/
func (c *MFECONF) PostQuery(urlStr string, reqBody string) (*Response, error) {
	reqBodyBytes, err := encrypt(reqBody, c)
	if err != nil {
		return nil, err
	}
	res, err := DoPostReq(urlStr, reqBodyBytes, c)
	if err != nil {
		return nil, err
	}
	res.Data, err = decrypt(res.Data, c)
	if err != nil {
		return nil, err
	}
	return res, nil
}
