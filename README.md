# gin-Middleware
## cat-go在gin中使用的中单件
#### 初始化
```
    r := gin.Default()
	r.Use(middleware.Cat(&cat.Options{
		AppId:      "cat-demo",
		Port:       2280,
		HttpPort:   8080,
		ServerAddr: "127.0.0.1",
	}))
```

### http(可按示例封装)
```
middleware.HttpGet(cxt, "http://localhost:8084/test")
```

#### gorm 
```
middleware.WithContextDB(cxt).Table("test").Count(&total)
```

#### redis 
```
cache := middleware.Cache().WithContext(cxt)
```