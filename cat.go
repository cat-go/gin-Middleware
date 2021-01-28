package middleware

import (
	"github.com/cat-go/cat"
	"github.com/cat-go/cat/message"
	"github.com/gin-gonic/gin"
)

const (
	CatCtx          = "cat_ctx"
	CatCtxRootTran  = "cat_root_tran"
	CatCtxRedisTran = "cat_redis_tran"
	CatCtxMysqlTran = "cat_mysql_tran"


)

//监控与链路追踪
func Cat(opts *cat.Options) gin.HandlerFunc {
	cat.DebugOn()
	cat.Init(opts)
	return func(c *gin.Context) {
		tran := cat.NewTransaction(cat.TypeUrl, c.Request.URL.Path)
		setTraceId(c, tran)
		tran.LogEvent(cat.TypeUrlMethod, c.Request.Method, c.FullPath())
		tran.LogEvent(cat.TypeUrlClient, c.ClientIP())
		c.Set(CatCtxRootTran, tran)
		defer func() {
			tran.Complete()
		}()
		c.Next()
	}
}

func setTraceId(c *gin.Context, tran message.Transactor) {
	var root, parent, child string
	root = c.Request.Header.Get(cat.RootId)
	parent = c.Request.Header.Get(cat.ParentId)
	child = c.Request.Header.Get(cat.ChildId)
	if root == "" {
		root = cat.MessageId()
	}
	if parent == "" {
		parent = root
	}
	if child == "" {
		child = cat.MessageId()
	}
	tran.SetRootMessageId(root)
	tran.SetParentMessageId(parent)
	tran.SetMessageId(child)
}
