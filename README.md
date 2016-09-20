kbang
===================================
http 压力测试工具，支持不同类型请求并发，支持请求权重

###Install package

```$ go get github.com/kaimixu/kbang```

### 使用方法
#####方法1(单请求并发)

```$ kbang <url>```
   
#####方法2(多请求并发)
```$ kbang -f kbang.conf```
  多请求并发需编写配置文件

### 配置文件描述
```
  # 每个请求配置需包含在单独的[request]配置块中
  [request]
  weight = 1
  # only support GET、POST
  method = GET
  url = http://www.baidu.com/
  
  [request]
  weight = 2
  method = POST
  content_type = text/plain
  url = http://www.baidu.com/
  post_data = a=1&b=2
  ```
