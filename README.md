# 轻量反向代理服务器
单awdp可以只下载v1版本。目前v1版本就是专门为了awdp的web设计的。
```cmd
Commands:
  proxy      简单反向代理
        --server_port  设置监听端口(必须，否则无法启动)
        --proxy_port   设置代理端口(必须，否则无法启动)
        --proxy_ip  	  设置代理ip(非必须，默认为127.0.0.1)
        --log_mode     设置日志模式默认 cli(仅仅控制台打印) save(保存到指定路径文件)
        --log_path     设置日志文件路径(如果log_mode为save则必须要填，反之则一定不要填写)
        --waf_rules    设置waf规则文件路径(非必须，不填写则不启动waf功能)
  webServer	   功能不齐暂时不支持

  start 开启多域名模式。（占用80 和 443端口，多域名模式暂时不支持waf功能）
       	--debug true 开启后会输出错误日志
  creat 创建一个域名网站。
		--website 网站名称（随意填写）
		--domain  域名（必填）
		--proxyUrl 反代地址（必填）
		--ssl 是否开启ssl（默认为false）
       
  -V, --version   Show the CLI version
  -h, --help      Show this help message
```

waf_rules

```yaml
low:
  allow:
    agent:
      - ".*"
    body:
      - ".*"
    url:
      - ".*"
  disallow:
    agent:
      - "curl"
    body:
      - "test"
      - "malicious"
    url:
      - "/api/restricted"
high:
  disallow:
    agent:
    body:
    url:
  allow:
    agent:
    url:
    body:
      - "testhight"
```

***waf高级规则***

***当low中disallow `关键词时` 但是high中allow包含`包含关键词的规则`按照high中allow***

***log日志规则***

***cli模式仅打印日志，save模式一定要指定log_path的路径***


