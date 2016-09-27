kbang
===================================
http 压力测试工具，支持不同类型请求并发，支持请求权重

###工具安装

```$ go get github.com/kaimixu/kbang```

### 使用方法
#####方法1(单请求并发)

```
   Usage: kbang [options...] <url>

options:
    -n  Number of requests to run (default: 10)
    -c  Number of requests to run concurrency (default: 1)
    -t  Request connection timeout in second (default: 1s)
    -H  Http header, eg. -H "Host: www.example.com"
    -k[=true|false]  Http keep-alive (default: false)
    -d  Http request body to POST
    -T  Content-type header to POST, eg. 'application/x-www-form-urlencoded'
        (Default：text/plain)
```
   
#####方法2(多请求并发)
```
   Usage: kbang [options...] -f kbang.conf
   ```
  多请求并发需编写配置文件

### 配置文件描述
```
  # 多请求配置文件，[request]用于区分不同请求，
  # weight表示请求权重，如下两请求权重比例为1:2,假如总请求数为300(-n 参数指定)，
  # 请求1执行100次，请求2执行200次。
  [request]
  weight = 1
  # only support GET、POST
  method = GET
  url = http://www.example.com/
  
  [request]
  weight = 2
  method = POST
  content_type = text/plain
  url = http://www.example.com/
  post_data = a=1&b=2
  ```
