package middleware

import (
	"fmt"
	"github.com/cat-go/cat"
	"github.com/cat-go/cat/message"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Header map[string]string

func HttpGet(c *gin.Context, url string) string {

	tran := cat.NewTransaction(cat.TypeUrlClient, url)
	defer tran.Complete()

	if t, ok := c.Get(CatCtxRootTran); ok {
		if t1, ok := t.(*message.Transaction); ok {
			fmt.Println(t1.GetRootMessageId(), t1.GetParentMessageId(), t1.GetMessageId(), "t1")
			cat.SetChildTraceId(t1, tran)
			t1.AddChild(tran)
		}
	}
	childId := cat.MessageId()
	tran.LogEvent(cat.TypeRemoteCall, "", cat.SUCCESS, childId)
	header := Header{
		cat.RootId: tran.GetRootMessageId(),
		cat.ParentId:  tran.GetMessageId(),
		cat.ChildId:   childId,
	}
	req, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		fmt.Println(err)
		return ""
	}

	for k, v := range header {
		req.Header.Add(k, v)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Add("Content-Type", "application/json")
	}

	clt := http.Client{
		Timeout: 30 * time.Second, //请求超时时间
	}
	res, err := clt.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}
