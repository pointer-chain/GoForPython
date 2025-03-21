// // main.go
package GoNeetHttp

//
///*
//#include <stdlib.h> // 引入C标准库，用于内存管理
//*/
//import "C" // 必须单独导入C包
//import (
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"io"
//	"net/http"
//	"net/url"
//	"strings"
//	"unsafe"
//)
//
//// 定义标准错误代码（包内常量）
//const (
//	ErrInvalidMethod    = 4001 // 非法HTTP方法
//	ErrHeaderParse      = 4002 // 请求头解析失败
//	ErrMissingUserAgent = 4003 // 缺少User-Agent
//	ErrProxyConfig      = 4004 // 代理配置错误
//	ErrBodySize         = 4005 // 代理配置错误
//	ErrRedirectExceed   = 3001 // 重定向次数超限
//	ErrNetwork          = 5001 // 网络请求失败
//	ErrReadResponse     = 5002 // 响应读取失败
//)
//
//// FreeCString 释放C语言字符串内存
//// 参数 cs: 需要释放的C字符串指针
////
////export FreeCString
//func FreeCString(cs *C.char) {
//	C.free(unsafe.Pointer(cs))
//}
//
//// PostUrlWithProxy 通过代理发起HTTP请求的C导出函数
//// 参数:
////
////	cMethod:          HTTP方法字符串指针 (C.char*)，仅接受GET/POST
////	cGetUrl:          目标URL字符串指针 (C.char*)
////	cHeaders:         JSON格式请求头字符串指针 (C.char*)
////	cProxyUrl:        代理地址字符串指针 (C.char*)，格式为scheme://host:port
////	cDisableRedirect: 禁用重定向标识指针 (C.char*)，"true"表示禁用
////
//// 返回值:
////
////	*C.char: 返回JSON格式的响应数据指针，需使用FreeCString释放
////
//// 安全注意事项:
////  1. 强制验证User-Agent头
////  2. 自动移除Authorization头
////  3. 限制响应体最大读取5MB
////  4. 白名单控制HTTP方法
////
////export PostUrlWithProxy
//func PostUrlWithProxy(cMethod, cGetUrl, cHeaders, cProxyUrl, cDisableRedirect, cBody *C.char) *C.char {
//	// 转换C字符串到Go字符串
//	method := strings.ToUpper(C.GoString(cMethod))
//	getUrl := C.GoString(cGetUrl)
//	cookie := C.GoString(cHeaders)
//	proxyUrl := C.GoString(cProxyUrl)
//	disableRedirect := C.GoString(cDisableRedirect) == "true"
//	bodyData := []byte(C.GoString(cBody))
//	var bodyReader io.Reader
//	// HTTP方法白名单验证
//	validMethods := map[string]bool{
//		"GET":    true,  // 允许GET
//		"POST":   true,  // 允许POST
//		"PUT":    false, // 禁用PUT
//		"DELETE": false, // 禁用DELETE
//		"PATCH":  false, // 禁用PATCH
//		"HEAD":   false, // 禁用HEAD
//	}
//	if !validMethods[method] {
//		return resultToC(nil, fmt.Errorf("无效的HTTP方法: %s", method))
//	}
//	// 解析headers JSON
//	var headers map[string]string
//	if err := json.Unmarshal([]byte(cookie), &headers); err != nil {
//		return resultToC(nil, fmt.Errorf("headers参数解析失败: %v", err))
//	}
//	// 在解析headers后增加必要字段校验
//	if _, ok := headers["User-Agent"]; !ok {
//		return resultToC(nil, fmt.Errorf("必须提供User-Agent请求头"))
//	}
//	// 防止意外泄露(敏感头字段)
//	var sensitiveHeaders = map[string]bool{
//		"Authorization":       true,
//		"Cookie":              true,
//		"Proxy-Authorization": true,
//	}
//	for h := range sensitiveHeaders {
//		delete(headers, h)
//	}
//	if contentType, ok := headers["Content-Type"]; ok && contentType == "application/x-www-form-urlencoded" {
//		formData, err := url.ParseQuery(string(bodyData))
//		if err != nil {
//			return resultToC(nil, fmt.Errorf("表单数据解析失败: %v", err))
//		}
//		bodyReader = strings.NewReader(formData.Encode())
//	} else {
//		bodyReader = bytes.NewReader(bodyData)
//	}
//	// 创建HTTP请求对象
//	req, err := http.NewRequest(method, getUrl, bodyReader)
//	if err != nil {
//		return resultToC(nil, err)
//	}
//	// 设置请求头
//	for key, value := range headers {
//		req.Header.Add(key, value)
//	}
//	// 代理配置处理（方案优先级）
//	// 1. 当提供有效代理地址时：创建带代理的自定义Transport
//	// 2. 无代理时：克隆默认Transport保证线程安全
//	var transport *http.Transport
//	if proxyUrl != "" {
//		// 解析代理地址
//		proxyURL, errProxy := url.Parse(proxyUrl)
//		if errProxy != nil {
//			return resultToC(nil, fmt.Errorf("代理地址解析失败: %v", errProxy))
//		}
//		// 创建带代理的传输层
//		transport = &http.Transport{
//			Proxy: http.ProxyURL(proxyURL),
//		}
//	} else {
//		// 使用默认传输层并克隆配置
//		transport = http.DefaultTransport.(*http.Transport).Clone()
//	}
//	// 创建HTTP客户端并禁止重定向
//	client := &http.Client{
//		Transport:     transport,
//		CheckRedirect: createRedirectPolicy(disableRedirect),
//	}
//	// 发送HTTP请求
//	res, err := client.Do(req)
//	if err != nil {
//		return resultToC(nil, err)
//	}
//	defer func(Body io.ReadCloser) {
//		// 确保关闭响应体
//		if err2 := Body.Close(); err2 != nil {
//			fmt.Printf("关闭响应体失败: %v\n", err2)
//		}
//	}(res.Body)
//	// // 安全读取响应体（限制最大5MB）
//	maxBodySize := 1024 * 1024 * 5 // 5MB
//	// 使用LimitReader防止内存溢出
//	bodyBytes, errRead := io.ReadAll(io.LimitReader(res.Body, int64(maxBodySize)))
//	if errRead != nil {
//		return resultToC(nil, fmt.Errorf("读取响应体失败: %v", errRead))
//	}
//	// 构造返回数据结构
//	result := map[string]interface{}{
//		"status":         res.Status,                     // 完整状态字符串（如"200 OK"）
//		"status_code":    res.StatusCode,                 // 状态码（如200）
//		"protocol":       res.Proto,                      // 协议版本（如HTTP/1.1）
//		"headers":        convertHeaders(res.Header),     // 响应头
//		"content_length": res.ContentLength,              // 声明的响应体长度
//		"body_size":      len(bodyBytes),                 // 实际读取的字节数
//		"cookies":        convertCookies(res.Cookies()),  // Cookies
//		"server":         res.Header.Get("Server"),       // 服务器信息
//		"content_type":   res.Header.Get("Content-Type"), // 内容类型
//		"date":           res.Header.Get("Date"),         // 响应日期
//		"body":           string(bodyBytes),              // 响应体内容
//		"redirects":      getRedirectHistory(res),        // 重定向历史
//	}
//
//	return resultToC(result, nil)
//}
//
//// resultToC 统一封装API响应格式
//// 设计规范：
//// - 成功时返回 {success:true, result:data}
//// - 失败时返回 {success:false, error:"message", error_code:num}
//// 错误代码体系：
//// 3000系列：重定向相关错误
//// 4000系列：客户端参数错误
//// 5000系列：服务端/网络错误
//func resultToC(data interface{}, err error) *C.char {
//	result := map[string]interface{}{
//		"success": err == nil,
//		"error":   nil,
//		"result":  data,
//	}
//
//	if err != nil {
//		// 错误处理
//		result["error"] = err.Error()
//		// 添加错误代码分类
//		// 使用类型断言和错误匹配进行精确判断
//		switch {
//		case strings.Contains(err.Error(), "无效的HTTP方法"):
//			result["error_code"] = ErrInvalidMethod
//		case strings.Contains(err.Error(), "headers参数解析"):
//			result["error_code"] = ErrHeaderParse
//		case strings.Contains(err.Error(), "必须提供User-Agent"):
//			result["error_code"] = ErrMissingUserAgent
//		case strings.Contains(err.Error(), "代理地址解析失败"):
//			result["error_code"] = ErrProxyConfig
//		case strings.Contains(err.Error(), "stopped after"):
//			result["error_code"] = ErrRedirectExceed
//		case strings.Contains(err.Error(), "读取响应体失败"),
//			strings.Contains(err.Error(), "stream error"):
//			result["error_code"] = ErrReadResponse
//		case strings.Contains(err.Error(), "body size exceeds"):
//			result["error_code"] = ErrBodySize
//		default:
//			// 网络相关错误的兜底判断
//			if isNetworkError(err) {
//				result["error_code"] = ErrNetwork
//			} else {
//				result["error_code"] = 5000 // 未知错误
//			}
//		}
//	}
//	// 序列化为JSON
//	jsonData, _ := json.Marshal(result)
//	// 转换为C字符串（需在调用端释放）
//	return C.CString(string(jsonData))
//}
//
//// convertHeaders 转换HTTP头到字典格式
//// 参数 h: http.Header类型
//// 返回值: 简化后的字典（只取每个头的第一个值）
//func convertHeaders(h http.Header) map[string][]string {
//	headers := make(map[string][]string)
//	for k, v := range h {
//		headers[k] = v
//	}
//	return headers
//}
//
//// convertCookies 转换http.Cookie为序列化友好的字典格式
//// 转换规则：
////  1. 仅保留Name/Value/Domain/Path四个字段
////  2. 自动过滤HttpOnly等敏感属性
////
//// 返回值示例：
////
////	[{"name": "session", "value": "abc123", "domain": ".example.com", "path": "/"}]
//func convertCookies(cookies []*http.Cookie) []map[string]string {
//	var result []map[string]string
//	for _, c := range cookies {
//		result = append(result, map[string]string{
//			"name":   c.Name,
//			"value":  c.Value,
//			"domain": c.Domain,
//			"path":   c.Path,
//		})
//	}
//	return result
//}
//
//// getRedirectHistory 获取重定向历史记录
//// 参数 res: 最终响应对象
//// 返回值: 历史URL列表（倒序，最新请求在前）
//func getRedirectHistory(res *http.Response) []string {
//	var urls []string
//	for res != nil {
//		urls = append(urls, res.Request.URL.String())
//		res = res.Request.Response
//	}
//	return urls
//}
//
//// createRedirectPolicy 创建HTTP重定向策略
//// 参数:
////
////	disable - true: 完全禁用重定向
////	         false: 启用重定向并限制最大5次跳转
////
//// 实现特点：
//// - 禁用时直接返回ErrUseLastResponse
//// - 启用时自动跟踪跳转链，防止重定向风暴
//func createRedirectPolicy(disable bool) func(*http.Request, []*http.Request) error {
//	if disable {
//		return func(req *http.Request, via []*http.Request) error {
//			return http.ErrUseLastResponse
//		}
//	}
//	return func(req *http.Request, via []*http.Request) error {
//		// 默认重定向策略（可扩展添加更多控制逻辑）
//		if len(via) >= 5 { // 示例：添加最大重定向次数限制
//			return fmt.Errorf("stopped after 5 redirects")
//		}
//		return nil
//	}
//}
//
//// 辅助函数判断网络错误
//func isNetworkError(err error) bool {
//	_, ok := err.(interface{ Timeout() bool })
//	if ok {
//		return true
//	}
//	return strings.Contains(err.Error(), "connection") ||
//		strings.Contains(err.Error(), "request canceled")
//}
//
//// main 空主函数（CGO编译要求）
//func main() {}
