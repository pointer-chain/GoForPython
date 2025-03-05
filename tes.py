import ctypes

# 加载共享库
dll = ctypes.CDLL('./get_login.dll')

# 定义函数签名
PostUrlWithProxy = dll.PostUrlWithProxy
PostUrlWithProxy.argtypes = [
    ctypes.c_char_p,  # getUrl
    ctypes.c_char_p,  # cookie
    ctypes.c_char_p,  # proxyUrl
]
PostUrlWithProxy.restype = ctypes.POINTER(ctypes.c_char)  # 返回值为 C 字符串指针

# 定义输入参数
get_url = b"https://example.com"  # 目标 URL
cookie = b"your_cookie_here"       # Cookie 字符串
proxy_url = b"http://your.proxy:8080"  # 代理 URL

# 调用函数
result_ptr = PostUrlWithProxy(get_url, cookie, proxy_url)

# 将返回的 C 字符串转换为 Python 字符串
result = ctypes.string_at(result_ptr).decode('utf-8')

# 打印结果
print("Response:", result)

# 释放 C 字符串内存
dll.free(result_ptr)
