// // main.go
package main

//
///*
//#include <stdlib.h> // 引入C标准库，用于内存管理
//*/
//import "C" // 必须单独导入C包
//import (
//	"encoding/json"
//	"fmt"
//	"io"
//	"net/http"
//	"net/url"
//	"strings"
//	"unsafe"
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
//// PostUrlWithProxy 主处理函数（C导出函数）
//// 参数:
////
////	cMethod: HTTP方法 (GET/POST)
////	cGetUrl: 请求URL
////	cHeaders: JSON格式的请求头
////	cProxyUrl: 代理地址（可选）
////
//// 返回值: JSON格式的C字符串指针
////
////export PostUrlWithProxy
//func PostUrlWithProxy(cMethod, cGetUrl, cHeaders, cProxyUrl, cDisableRedirect *C.char) *C.char {
//	// 转换C字符串到Go字符串
//	method := strings.ToUpper(C.GoString(cMethod))
//	getUrl := C.GoString(cGetUrl)
//	cookie := C.GoString(cHeaders)
//	proxyUrl := C.GoString(cProxyUrl)
//	disableRedirect := C.GoString(cDisableRedirect) == "true"
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
//	delete(headers, "Authorization")
//
//	// 创建HTTP请求对象
//	req, err := http.NewRequest(method, getUrl, nil)
//	if err != nil {
//		return resultToC(nil, err)
//	}
//	// 设置请求头
//	for key, value := range headers {
//		req.Header.Add(key, value)
//	}
//	// 代理配置处理
//	var transport *http.Transport
//	if proxyUrl != "" {
//		// 解析代理地址
//		proxyURL, errProxy := url.Parse(proxyUrl)
//		if errProxy != nil {
//			return resultToC(nil, err)
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
//		Transport: transport,
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
//// 新增工厂函数
//func createRedirectPolicy(disable bool) func(*http.Request, []*http.Request) error {
//	if disable {
//		return func(_ *http.Request, _ []*http.Request) error {
//			return http.ErrUseLastResponse
//		}
//	}
//	return nil // 使用默认重定向策略
//}
//// resultToC 统一处理返回结果
//// 参数:
////
////	data: 成功时的返回数据
////	err: 错误信息
////
//// 返回值: JSON格式的C字符串指针
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
//		if strings.Contains(err.Error(), "headers参数解析") {
//			result["error_code"] = 4001
//		} else if strings.Contains(err.Error(), "网络请求失败") {
//			result["error_code"] = 5001
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
//// convertCookies 转换Cookies到字典列表
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
//// main 空主函数（CGO编译要求）
//func main() {}
