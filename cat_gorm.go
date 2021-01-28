package middleware

import (
	"context"
	"fmt"
	"github.com/cat-go/cat"
	"github.com/cat-go/cat/message"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

//将上下文应用到gorm实例
func WithCtxGorm(ctx context.Context, db *gorm.DB) *gorm.DB {
	if ctx == nil {
		return db
	}
	return db.Set(CatCtx, ctx)
}

// 添加gormCallback
func AddGormCallbacks(db *gorm.DB) {
	callbacks := newCallbacks()
	registerCallbacks(db, "create", callbacks)
	registerCallbacks(db, "query", callbacks)
	registerCallbacks(db, "update", callbacks)
	registerCallbacks(db, "delete", callbacks)
	registerCallbacks(db, "row_query", callbacks)
}

type callbacks struct{}

func newCallbacks() *callbacks {
	return &callbacks{}
}

func (c *callbacks) beforeCreate(scope *gorm.Scope)   { c.before(scope) }
func (c *callbacks) afterCreate(scope *gorm.Scope)    { c.after(scope, "INSERT") }
func (c *callbacks) beforeQuery(scope *gorm.Scope)    { c.before(scope) }
func (c *callbacks) afterQuery(scope *gorm.Scope)     { c.after(scope, "SELECT") }
func (c *callbacks) beforeUpdate(scope *gorm.Scope)   { c.before(scope) }
func (c *callbacks) afterUpdate(scope *gorm.Scope)    { c.after(scope, "UPDATE") }
func (c *callbacks) beforeDelete(scope *gorm.Scope)   { c.before(scope) }
func (c *callbacks) afterDelete(scope *gorm.Scope)    { c.after(scope, "DELETE") }
func (c *callbacks) beforeRowQuery(scope *gorm.Scope) { c.before(scope) }
func (c *callbacks) afterRowQuery(scope *gorm.Scope)  { c.after(scope, "QUERY") }

func (c *callbacks) before(scope *gorm.Scope) {
	catCtx, ok := scope.Get(CatCtx)
	if !ok {
		return
	}
	ctx, ok := catCtx.(*gin.Context)
	if !ok {
		return
	}
	tran := cat.NewTransaction(cat.TypeSql, scope.Dialect().GetName())
	ctx.Set(CatCtxMysqlTran, tran)
	if t, ok := ctx.Get(CatCtxRootTran); ok {
		if t1, ok := t.(message.Transactor); ok {
			cat.SetChildTraceId(t1, tran)
			t1.AddChild(tran)
		}
	}
}

func (c *callbacks) after(scope *gorm.Scope, operation string) {
	catCtx, ok := scope.Get(CatCtx)
	if !ok {
		return
	}
	ctx, ok := catCtx.(*gin.Context)
	if !ok {
		return
	}
	if tran, ok := ctx.Value(CatCtxMysqlTran).(message.Transactor); ok {
		tran.LogEvent(cat.TypeSqlOp, operation)
		tran.LogEvent(cat.TypeSqlVal, scope.SQL)
		tran.Complete()
	}
}

func registerCallbacks(db *gorm.DB, name string, c *callbacks) {
	beforeName := fmt.Sprintf("tracing:%v_before", name)
	afterName := fmt.Sprintf("tracing:%v_after", name)
	gormCallbackName := fmt.Sprintf("gorm:%v", name)
	switch name {
	case "create":
		db.Callback().Create().Before(gormCallbackName).Register(beforeName, c.beforeCreate)
		db.Callback().Create().After(gormCallbackName).Register(afterName, c.afterCreate)
	case "query":
		db.Callback().Query().Before(gormCallbackName).Register(beforeName, c.beforeQuery)
		db.Callback().Query().After(gormCallbackName).Register(afterName, c.afterQuery)
	case "update":
		db.Callback().Update().Before(gormCallbackName).Register(beforeName, c.beforeUpdate)
		db.Callback().Update().After(gormCallbackName).Register(afterName, c.afterUpdate)
	case "delete":
		db.Callback().Delete().Before(gormCallbackName).Register(beforeName, c.beforeDelete)
		db.Callback().Delete().After(gormCallbackName).Register(afterName, c.afterDelete)
	case "row_query":
		db.Callback().RowQuery().Before(gormCallbackName).Register(beforeName, c.beforeRowQuery)
		db.Callback().RowQuery().After(gormCallbackName).Register(afterName, c.afterRowQuery)
	}
}
