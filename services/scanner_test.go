package services

import (
	"net/http"
	"testing"
)

// newHeader 快速构造响应头
func newHeader(kv ...string) http.Header {
	h := http.Header{}
	for i := 0; i+1 < len(kv); i += 2 {
		h.Set(kv[i], kv[i+1])
	}
	return h
}

func contains(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

// 只配了 BodyKeys 的规则必须能通过 body 命中
// （回归测试：修复前这段匹配逻辑嵌在 Header 循环内，Headers 为空的规则全部失效）
func TestIdentifyFingerprint_BodyOnlyRules(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{"RuoYi", `<html><head><title>若依系统</title></head><body>/prod-api/captcha</body></html>`, "RuoYi"},
		{"ZhiyuanOA", `<html><body>致远OA A8</body></html>`, "ZhiyuanOA"},
		{"WeaverOA", `<html><body>/wui/theme/ecology</body></html>`, "WeaverOA"},
		{"Nacos", `<html><head><title>Nacos</title></head></html>`, "Nacos"},
		{"WordPress", `<html><body>wp-content/themes</body></html>`, "WordPress"},
		{"SwaggerUI", `<html><head><title>Swagger UI</title></head></html>`, "Swagger UI"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := identifyFingerprint(newHeader("Content-Type", "text/html"), c.body)
			if !contains(got, c.want) {
				t.Errorf("期望命中 %s，实际得到 %v", c.want, got)
			}
		})
	}
}

// 配了 Headers 的规则必须能通过 header 命中
func TestIdentifyFingerprint_HeaderRules(t *testing.T) {
	cases := []struct {
		name   string
		header http.Header
		want   string
	}{
		{"Nginx", newHeader("Server", "nginx/1.24.0"), "Nginx"},
		{"Shiro", newHeader("Set-Cookie", "rememberMe=deleteMe"), "Shiro"},
		{"Jenkins", newHeader("X-Jenkins", "2.441"), "Jenkins"},
		{"Tengine", newHeader("Server", "Tengine/2.3.3"), "Tengine"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := identifyFingerprint(c.header, "")
			if !contains(got, c.want) {
				t.Errorf("期望命中 %s，实际得到 %v", c.want, got)
			}
		})
	}
}

// Header 与 Body 双特征规则：header 未命中时应回退到 body 匹配
func TestIdentifyFingerprint_FallbackToBody(t *testing.T) {
	// JeecgBoot 同时配了 header 和 body，这里只给 body 特征
	got := identifyFingerprint(newHeader("Server", "nginx"), `<html><body>jeecg-boot 管理后台</body></html>`)
	if !contains(got, "JeecgBoot") {
		t.Errorf("期望通过 body 命中 JeecgBoot，实际得到 %v", got)
	}
}

// 无任何特征时不应产生误报
func TestIdentifyFingerprint_NoFalsePositive(t *testing.T) {
	got := identifyFingerprint(newHeader("Server", "Apache"), `<html><body>Hello World</body></html>`)
	if len(got) != 0 {
		t.Errorf("不应产生误报，实际得到 %v", got)
	}
}

// 一个页面命中多个指纹时应全部返回
func TestIdentifyFingerprint_MultipleMatches(t *testing.T) {
	got := identifyFingerprint(newHeader("Server", "nginx"), `<html><head><title>若依</title></head><body>wp-content</body></html>`)
	if !contains(got, "RuoYi") || !contains(got, "Nginx") || !contains(got, "WordPress") {
		t.Errorf("期望同时命中 RuoYi/Nginx/WordPress，实际得到 %v", got)
	}
}
