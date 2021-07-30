package logger

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	envA         = os.Getenv("FOFA_EMAIL")
	envK         = os.Getenv("FOFA_KEY")
	pWARN        = "\033[1;31m[!]\033[0m "
	pINFO        = "\033[1;33m[+]\033[0m "
	pDEBUG       = "\033[1;34m[-]\033[0m "
	pSUCCESS     = "\033[1;32m[*]\033[0m "
	isWin        = false
	usageCommand = "Usage:\n  ./fofa -m fofa_email -k fofa_key -q 'header=\"elastic\"' -s 10000 -o data.xlsx\n" +
		"  ./fofa -m fofa_email -k fofa_key -f query_rules_file.txt -s 10000 -o data.xlsx\n"
	usageOptions = `Options:
  -h, --help
  -m, --mail MAIL            fofa email account (default: ` + envA + `)
  -k, --key KEY              fofa api key (default: ` + envK + `)
  -q, --query QUERY          query string (default: '')
  -f, --file FILE            batch query rules file (default: '')
  -s, --size SIZE            export data volume (default: 10000)
  -o, --output OUTPUT        output filename / absolute path (default: data.xlsx)
  -g, --grammar              fofa search grammar help table
  -t, --tip TIP              fofa search keyword tip droplist
  -ih, --iconhash ICONHASH   calculate url specified favicon icon_hash       
`
	fofaGrammar = `
[+]                Rule                                   Mark               
 ------------------------------------------- -------------------------------- 
  domain="qq.com"                             搜索根域名带有qq.com的网站                
  host=".gov.cn"                              从host中搜索".gov.cn"               
  ip="1.1.1.1"                                从ip中搜索包含"1.1.1.1"的网站            
  ip="220.181.111.1/24"                       查询IP为"220.181.111.1"的C段资产       
  port="6379"                                 查找对应"6379"端口的资产                 
  title="beijing"                             从标题中搜索"北京"                      
  status_code="402"                           查询服务器状态为"402"的资产                
  protocol="quic"                             查询quic协议资产                      
  header="elastic"                            从http头中搜索"elastic"              
  body="网络空间测绘"                         从html正文中搜索"网络空间测绘"              
  os="centos"                                 搜索CentOS资产                      
  server=="Microsoft-IIS/10"                  搜索IIS10服务器                      
  app="Microsoft-Exchange"                    搜索Microsoft-Exchange设备          
  base_protocol="udp"                         搜索指定udp协议的资产                    
  banner=users && protocol=ftp                搜索FTP协议中带有users文本的资产            
  icp="京ICP证030173号"                       查找备案号为"京ICP证030173号"的网站         
  icon_hash="-247388890"                      搜索使用此icon的资产(VIP)               
  js_name="js/jquery.js"                      查找网站正文中包含js/jquery.js的资产        
  js_md5="82ac3f14327a8b7ba49baa208d4eaa15"   查找js源码与之匹配的资产                   
  type=service                                搜索所有协议资产，支持subdomain和service两种  
  is_domain=true                              搜索域名的资产                         
  ip_ports="80,161"                           搜索同时开放80和161端口的ip               
  port_size="6"                               查询开放端口数量等于"6"的资产(VIP)           
  port_size_gt="6"                            查询开放端口数量大于"6"的资产(VIP)           
  port_size_lt="12"                           查询开放端口数量小于"12"的资产(VIP)          
  is_ipv6=true                                搜索ipv6的资产                       
  is_fraud=false                              排除仿冒/欺诈数据                       
  is_honeypot=false                           排除蜜罐数据(VIP)                     
  country="CN"                                搜索指定国家(编码)的资产                   
  region="Xinjiang"                           搜索指定行政区的资产                      
  city="Ürümqi"                               搜索指定城市的资产                       
  asn="19551"                                 搜索指定asn的资产                      
  org="Amazon.com,Inc."                       搜索指定org(组织)的资产                  
  cert="baidu"                                搜索证书(https或者imaps等)中带有baidu的资产  
  cert.subject="Oracle Corporation"           搜索证书持有者是OracleCorporation的资产    
  cert.issuer="DigiCert"                      搜索证书颁发者为DigiCertInc的资产          
  cert.is_valid=true                          验证证书是否有效,true有效,false无效(VIP)    
  after="2017" && before="2017-10-01"         限定时间范围
  
[+] 高级搜索：可以使用括号 () / 和 && / 或 || / 完全匹配 == / 不为 != 等逻辑运算符。
`
)

const (
	VERSION     = 1.0
	asciiBanner = `
      ░░░░▐▐░░░  dMMMMMP .aMMMb  dMMMMMP .aMMMb 
 ▐  ░░░░░▄██▄▄  dMP     dMP"dMP dMP     dMP"dMP 
  ▀▀██████▀░░  dMMMP   dMP dMP dMMMP   dMMMMMP  
  ░░▐▐░░▐▐░░  dMP     dMP.aMP dMP     dMP dMP   
 ▒▒▒▐▐▒▒▐▐▒  dMP      VMMMP" dMP     dMP dMP
 https://github.com/inspiringz/fofa
`

	winAsciiBanner = "\n  ,-.       _,---._ __  / \\ \n" +
		" /  )    .-'       `./ /   \\ \n" +
		"(  (   ,'            `/    /|\n" +
		" \\  `-\"             \\'\\   / |\n" +
		"  `.              ,  \\ \\ /  |\n" +
		"   /`.          ,'-`----Y   |\n" +
		"  (            ;        |   '\n" +
		"  |  ,-.    ,-'         |  /\n" +
		"  |  | (   |            | /\n" +
		"  )  |  \\  `.___________|/\n" +
		"  `--'   `--'\n" +
		"https://github.com/inspiringz/fofa\n"
)

func InitPlatform() {
	if strings.HasSuffix(strings.ToLower(os.Args[0]), ".exe") {
		pWARN = "[!] "
		pINFO = "[+] "
		pSUCCESS = "[*] "
		pDEBUG = "[-] "
		isWin = true
	}
	usageCommand = strings.Replace(usageCommand, "./fofa", os.Args[0], 2)
	return
}

func randColor() string {
	rand.Seed(time.Now().UnixNano())
	colorNum := rand.Intn(6) + 31
	return strconv.Itoa(colorNum)
}

func AsciiBanner() {
	if isWin != true {
		fmt.Println("\033[1;" + randColor() + "m" + asciiBanner + "\033[0m")
	} else {
		fmt.Println(winAsciiBanner)
	}
}

func Usage() {
	fmt.Println(usageCommand)
	fmt.Println(usageOptions)
}

func FofaGrammar() {
	fmt.Println(fofaGrammar)
}

func Warn(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, pWARN+format+"\n", args...)
}

func Info(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, pINFO+format+"\n", args...)
}

func Success(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, pSUCCESS+format+"\n", args...)
}

func Debug(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, pDEBUG+format+"\n", args...)
}
