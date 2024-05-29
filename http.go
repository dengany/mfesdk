package mfesdk

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func doPostReq(urlStr string, reqBody []byte, cfg *MFECONF) (*Response, error) {
	ctx := context.Background()
	client := g.Client()
	if cfg.IsProd {
		urlStr = BASE_API_URL + urlStr
	} else {
		urlStr = BASE_TEST_API_URL + urlStr
	}
	Sign, _ := Sign(reqBody, cfg)
	timd := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("X-Security", "CFCA")
	client.SetHeader("X-AGENCY", cfg.AgencyNo)
	client.SetHeader("X-Sign", Sign)
	client.SetHeader("X-Time", timd)
	client.SetHeader("X-Trace", fmt.Sprint(time.Now().UnixNano()))
	req, _ := json.Marshal(g.Map{"param": string(reqBody)})
	res, err := client.Post(ctx, urlStr, req)
	if err != nil {
		g.Log().Error(ctx, err)
		return nil, err
	}
	defer res.Close()
	respSign := res.Header.Get("X-Sign")

	body := res.ReadAll()
	response := &Response{}
	// err = gconv.Scan(body, response)
	err = json.Unmarshal(body, response)
	if err != nil {
		g.Log().Error(ctx, err)
		return nil, err
	}
	if response.Code != "000000" {
		log.Println("请求返回失败 CODE:", response.Code, "请求返回失败 ERROR:", response.Message, "请求返回失败 DATA:", response.Data)
		return nil, fmt.Errorf(response.Message)
	}
	if ok, err := Verify([]byte(response.Data), []byte(respSign), cfg); !ok {
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
	reqBodyBytes, err := Encrypt(reqBody, c)
	if err != nil {
		return nil, err
	}
	res, err := doPostReq(urlStr, []byte(reqBodyBytes), c)
	if err != nil {
		return nil, err
	}
	res.Data, err = Decrypt(res.Data, c)
	if err != nil {
		return nil, err
	}
	return res, nil
}

/*
文件上传
*/
func (c *MFECONF) UploadFile(filepath string) (*Response, error) {

	res, err := doUploadFile("/file/api/upload", filepath, c)
	if err != nil {
		return nil, err
	}
	res.Data, err = Decrypt(res.Data, c)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func doUploadFile(urlStr string, filePath string, cfg *MFECONF) (*Response, error) {
	// ctx := gctx.New()

	if cfg.IsProd {
		urlStr = BASE_API_URL + urlStr
	} else {
		urlStr = BASE_TEST_API_URL + urlStr
	}
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		panic(err)
	}
	hashPlains := hex.EncodeToString(hash.Sum(nil))
	// fmt.Println("Request multipart file hash plains:", hashPlains)

	hashCipher, err := fileencrypt(hashPlains, cfg.PublicKey)
	if err != nil {
		panic(err)
	}
	x_hash := base64.StdEncoding.EncodeToString(hashCipher)
	// fmt.Println("Request multipart file hash cipher:", x_hash)

	signature, err := Sign([]byte(x_hash), cfg)
	if err != nil {
		panic(err)
	}
	requestSign := signature
	// fmt.Println("Request sign:", requestSign)

	requestTime := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	requestTrace := fmt.Sprint(time.Now().UnixNano())

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("hash", hashPlains)

	file, err = os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		log.Fatal(err)
	}
	if _, err := io.Copy(part, file); err != nil {
		log.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", urlStr, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Agency", cfg.AgencyNo)
	req.Header.Set("X-Security", "CFCA")
	req.Header.Set("X-Sign", requestSign)
	req.Header.Set("X-Hash", x_hash)
	req.Header.Set("X-Time", requestTime)
	req.Header.Set("X-Trace", requestTrace)

	// fmt.Println("Request header:", req.Header)

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	fmt.Println("Response HTTP status:", res.Status)

	resbody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	response := &Response{}
	err = json.Unmarshal(resbody, response)
	// err = gconv.Scan(resbody, response)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	// if response.Code != "000000" {
	log.Println("请求返回失败 CODE:", response.Code, "请求返回失败 ERROR:", response.Message, "请求返回失败 DATA:", response.Data)
	// 	return nil, fmt.Errorf(response.Message)
	// }
	return response, nil
}
