package cli

import (
	"errors"
	"fmt"
	"strings"
)

func Run(args []string) error {
	if len(args) == 0 {
		return errors.New("no arguments provided. Use --help for usage")
	}

	switch args[0] {
	case "-V", "--version":
		fmt.Println("WebProxy CLI version 1.0.0")
	case "-h", "--help":
		printHelp()
	default:
		// 调用子命令
		return executeCommand(args)
	}
	return nil
}

func printHelp() {
	fmt.Println(`WebProxy CLI
Commands:
  proxy      Run the proxy server
       --server_port  设置监听端口(必须，否则无法启动)
       --proxy_port   设置代理端口(必须，否则无法启动)
       --proxy_ip  	  设置代理ip(非必须，默认为127.0.0.1)
       --log_mode     设置日志模式默认 cli(仅仅控制台打印) save(保存到指定路径文件)
       --log_path     设置日志文件路径(如果log_mode为save则必须要填，反之则一定不要填写)
       --waf_rules    设置waf规则文件路径(非必须，不填写则不启动waf功能)
  -V, --version   Show the CLI version
  -h, --help      Show this help message`)
}

func executeCommand(args []string) error {
	command := args[0]
	parsedArgs := parseArgs(args[1:])

	switch command {
	case "proxy":
		return runProxy(parsedArgs)
	default:
		return fmt.Errorf("unknown command: %s. Use --help for usage", command)
	}
}

// parseArgs 将参数解析为键值对
func parseArgs(args []string) map[string]string {
	parsed := make(map[string]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			key := strings.TrimPrefix(arg, "--")
			// 检查是否有下一个值，并且下一个值不是另一个选项
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				parsed[key] = args[i+1]
				i++
			} else {
				parsed[key] = "" // 没有值的选项
			}
		} else {
			// 非选项参数，按顺序存储
			parsed[arg] = ""
		}
	}
	return parsed
}
