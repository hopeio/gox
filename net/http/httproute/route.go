// Package httproute 把请求路径收敛成适合做指标标签的路由模板。
package httproute

import (
	"net/http"
	"strings"
)

// Of 优先用 ServeMux 匹配到的模板，否则按段收敛路径。
func Of(r *http.Request) string {
	if p := r.Pattern; p != "" {
		if _, route, ok := strings.Cut(p, " "); ok {
			return route
		}
		return p
	}
	return Collapse(r.URL.Path)
}

// Collapse 把 id/md5 之类的可变段换成 `:v`，防止指标基数爆炸。
func Collapse(p string) string {
	if p == "" {
		return "/"
	}
	segs := strings.Split(p, "/")
	for i, s := range segs {
		if looksVariable(s) {
			segs[i] = ":v"
		}
	}
	return strings.Join(segs, "/")
}

// looksVariable：纯数字，或较长的十六进制（md5/sha/雪花 id）。
func looksVariable(s string) bool {
	if s == "" {
		return false
	}
	digits, hex := true, true
	for _, r := range s {
		isDigit := r >= '0' && r <= '9'
		if !isDigit {
			digits = false
		}
		if !isDigit && !(r >= 'a' && r <= 'f') && !(r >= 'A' && r <= 'F') {
			hex = false
		}
	}
	if digits {
		return true
	}
	return hex && len(s) >= 16
}
