package services

import (
	"io"
	"net/http"
	"strings"
	"time"
)

// Fingerprint 代表一条指纹规则
type Fingerprint struct {
	Name     string   // 指纹名称（例如 "RuoYi"）
	Headers  []string // 匹配响应头特征（例如 "Set-Cookie: rememberMe"）
	BodyKeys []string // 匹配网页正文特征（例如 "若依", "ruoyi"）

}

// 预定义规则库（今天先写几条国内常见的）
var FingerprintRules = []Fingerprint{
	// ======== OA 与 国内管理系统 ========
	{
		Name:     "RuoYi",
		Headers:  []string{},
		BodyKeys: []string{"<title>若依", "/prod-api/", "ruoyi"}, // 加上 <title> 降低误报
	},
	{
		Name:     "JeecgBoot",
		Headers:  []string{"X-Powered-By: Jeecg-Boot"},
		BodyKeys: []string{"jeecg-boot", "JeecgBoot", "积木报表"},
	},
	{
		Name:     "ZhiyuanOA",
		Headers:  []string{},
		BodyKeys: []string{"致远OA", "seeyon", "/seeyon/"},
	},
	{
		Name:     "WeaverOA",
		Headers:  []string{},
		BodyKeys: []string{"泛微", "weaver", "/wui/"},
	},
	{
		Name:     "Nginx",
		Headers:  []string{"Server: nginx"},
		BodyKeys: []string{},
	},
	{
		Name:     "Tengine",
		Headers:  []string{"Server: Tengine"},
		BodyKeys: []string{},
	},

	// ======== Java 生态与框架 ========
	{
		Name:     "Shiro",
		Headers:  []string{"Set-Cookie: rememberMe=deleteMe"}, // 强特征
		BodyKeys: []string{},
	},
	{
		Name:     "Spring Boot",
		Headers:  []string{"X-Application-Context"},
		BodyKeys: []string{"Whitelabel Error Page", "Spring Boot"},
	},
	{
		Name:     "Swagger UI",
		Headers:  []string{},
		BodyKeys: []string{"<title>Swagger UI</title>", "swagger-ui"},
	},
	{
		Name:     "Druid",
		Headers:  []string{},
		BodyKeys: []string{"<title>Druid Stat Index</title>", "druid.index"},
	},

	// ======== 中间件与基础设施 ========
	{
		Name:     "Nacos",
		Headers:  []string{},
		BodyKeys: []string{"<title>Nacos</title>", "console-ui/public/img/favicon.ico"}, // 收窄特征
	},
	{
		Name:     "Tomcat",
		Headers:  []string{"Server: Apache-Coyote"},
		BodyKeys: []string{"Apache Tomcat"},
	},
	{
		Name:     "Jenkins",
		Headers:  []string{"X-Jenkins:"}, // 注意冒号不能省： SplitN 按冒号切分，缺冒号整条规则会被跳过（修复前 Jenkins 从未生效）
		BodyKeys: []string{"<title>Dashboard [Jenkins]</title>"},
	},
	{
		Name:     "GitLab",
		Headers:  []string{},
		BodyKeys: []string{"<title>GitLab</title>", "GitLab Community Edition"},
	},
	{
		Name:     "Grafana",
		Headers:  []string{},
		BodyKeys: []string{"<title>Grafana</title>"},
	},
	{
		Name:     "Elasticsearch",
		Headers:  []string{"X-Elastic-Product: Elasticsearch"},
		BodyKeys: []string{"You Know, for Search", "\"cluster_name\""},
	},
	{
		Name:     "Kibana",
		Headers:  []string{},
		BodyKeys: []string{"<title>Kibana</title>"},
	},
	{
		Name:     "Zabbix",
		Headers:  []string{},
		BodyKeys: []string{"<title>Zabbix</title>"},
	},
	{
		Name:     "宝塔面板",
		Headers:  []string{},
		BodyKeys: []string{"<title>宝塔面板</title>", "bt.cn", "btwaf"},
	},

	// ======== 常见 Web 应用与语言 ========
	{
		Name:     "WordPress",
		Headers:  []string{},
		BodyKeys: []string{"wp-content", "wp-includes", "wordpress"},
	},
	{
		Name:     "ThinkPHP",
		Headers:  []string{"X-Powered-By: ThinkPHP"},
		BodyKeys: []string{"ThinkPHP"},
	},
	{
		Name:     "Laravel",
		Headers:  []string{"Set-Cookie: laravel_session"},
		BodyKeys: []string{},
	},
	{
		Name:     "PhpMyAdmin",
		Headers:  []string{},
		BodyKeys: []string{"<title>phpMyAdmin</title>", "pma_username"},
	},
	{
		Name:     "Discuz",
		Headers:  []string{},
		BodyKeys: []string{"Discuz!", "discuz"},
	},
}

// identifyFingerprint 接收响应数据，返回匹配到的指纹名称列表
// 匹配策略：任意一条 Header 规则命中，或任意一条 Body 关键词命中，即判定为该指纹
func identifyFingerprint(headers map[string][]string, body string) []string {
	var matches []string

	// 预处理：body 统一转小写，避免每条规则、每个关键词重复转换
	lowerBody := strings.ToLower(body)

	for _, rule := range FingerprintRules {
		isMatched := false

		//1. 匹配 Header（任一 Header 规则命中即可）
		for _, headerRule := range rule.Headers {
			parts := strings.SplitN(headerRule, ":", 2)
			if len(parts) != 2 {
				continue // 规则格式不含冒号，跳过该条
			}
			headerKey := strings.TrimSpace(parts[0])
			headerVal := strings.ToLower(strings.TrimSpace(parts[1]))
			// 遍历响应的 headers 看是否包含规则里的值
			for respKey, respVals := range headers {
				if !strings.EqualFold(respKey, headerKey) {
					continue
				}
				for _, val := range respVals {
					if strings.Contains(strings.ToLower(val), headerVal) {
						isMatched = true
						break
					}
				}
				if isMatched {
					break
				}
			}
			if isMatched {
				break
			}
		}

		// 2. Header 未命中时，再匹配 Body 关键词（任一命中即可）
		// ⚠️ 历史 Bug：这段代码曾经写在上面 Header 循环的内部，
		// 导致 Headers 为空的规则（RuoYi、致远OA、泛微、Nacos、WordPress 等 14 条）
		// 的 BodyKeys 永远不会被匹配，指纹引擎半数规则失效。已修复，勿改回去。
		if !isMatched {
			for _, keyword := range rule.BodyKeys {
				if strings.Contains(lowerBody, strings.ToLower(keyword)) {
					isMatched = true
					break
				}
			}
		}

		if isMatched {
			matches = append(matches, rule.Name)
		}
	}
	return matches
}

func Probe(client *http.Client, url string, retries int) (int, string, string, string, error) {
	var resp *http.Response
	var err error

	// 自定义 User-Agent，伪装成 Chrome 浏览器，绕过基础 WAF 拦截
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	// 循环重试逻辑
	for i := 0; i < retries; i++ {
		// 1. 必须用 http.NewRequest 才能自定义请求头
		req, reqErr := http.NewRequest("GET", url, nil)
		if reqErr != nil {
			return 0, "", "", "", reqErr // URL 本身格式错误，重试也没用，直接返回
		}
		//2.穿上马甲
		req.Header.Set("User-Agent", userAgent)
		//3.发送请求
		resp, err = client.Do(req)

		// 如果请求成功且状态码小于 500（如 200、403），说明服务器已正常响应，直接跳出重试循环
		if err == nil && resp.StatusCode < 500 {
			break
		}

		// 准备重试前，必须关闭上一次请求的响应体，防止内存泄漏
		if resp != nil {
			resp.Body.Close() // 极其重要！重试前必须关闭上一次的 Body，否则内存泄漏
		}

		// 没到最后一次重试，歇 1 秒再试
		if i < retries-1 {
			time.Sleep(1 * time.Second)
		}
	}

	// 如果重试完了依然报错，返回最后的错误
	if err != nil {
		return 0, "", "", "", err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	server := resp.Header.Get("Server")
	if server == "" {
		server = "-" // 没拿到就显示横杠，保持输出整齐
	}

	// 读取 Body 内容，限制为前 1MB，防止读取巨型文件导致 OOM（内存溢出）
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return statusCode, server, "-", "-", err
	}

	// 提取网页标题 <title>
	bodyStr := string(bodyBytes)
	title := "-"
	titleStart := strings.Index(bodyStr, "<title>")
	titleEnd := strings.Index(bodyStr, "</title>")

	if titleStart != -1 && titleEnd != -1 && titleEnd > titleStart {
		title = strings.TrimSpace(bodyStr[titleStart+7 : titleEnd])
		// 清洗标题中的换行符
		title = strings.ReplaceAll(title, "\n", " ")
		title = strings.ReplaceAll(title, "\r", " ")
		// 截断过长的标题，防止终端输出爆炸
		if len(title) > 30 {
			title = title[:30] + "..."
		}
	}

	// 3. 【新增】调用指纹识别引擎
	fingerprints := identifyFingerprint(resp.Header, bodyStr)
	fpStr := strings.Join(fingerprints, "|")
	if fpStr == "" {
		fpStr = "-"
	}
	return statusCode, server, title, fpStr, nil
}
