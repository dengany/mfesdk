package main

import (
	"fmt"

	mfesdk "github.com/dengany/mfesdk"

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
