# MFE SDK

This is the MFE SDK.

这是现代金控 api 接口调用的SDK封装。
已经完成了公钥私钥读取，签名验证，请求加密，响应解密，请求参数封装，响应结果解析等功能。

## 安装

```
go get -u github.com/dengany/mfesdk

## 简单示例

```go
package main

import (
	"fmt"

	"github.com/dengany/mfesdk"

	"github.com/gogf/gf/v2/encoding/gjson"
)

type Demo struct {
	Name string `json:"name"`
}

func main() {
	op := &mfesdk.MfeOption{
		PubPath:  "cer/xxxx.cer",
		PriPath:  "cer/xxx.pfx",
		PriPwd:   "123456",
		AgencyNo: "20240226175310079X",
		IsProd:   false,
	}
	mfe := mfesdk.NewMfe(op)
	params := gjson.MustEncodeString(Demo{Name: "1234"})
	ss, err := mfe.PostQuery(mfesdk.Demo_API, params)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ss)
}
