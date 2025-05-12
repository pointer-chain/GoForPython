# -*- coding=utf-8 -*-
"""
@Time ： 2025/02/17 09:06
@Auth ： Yang
@File ：lib_proxy.py
"""
import json
import ctypes
import platform
import os
import subprocess


class LibProxy:
    def __init__(self, ctx):
        self.context = ctx
        # 获取系统架构信息
        architecture = platform.architecture()
        # 提取位数信息
        bits = architecture[0]
        print(f"当前系统位数是: {bits}")
        system = platform.system()
        if system == 'Windows':
            lib_proxy = os.path.join(os.getcwd(), "libproxy5.dll")
        elif system == 'Linux':
            lib_proxy = os.path.join(os.getcwd(), "libproxy5.so")
            try:
                subprocess.run(['chmod', '+r', lib_proxy], check=True)
                print(f"成功修改 {lib_proxy} 的权限")
            except subprocess.CalledProcessError as e:
                print(f"修改权限时出错: {e}")
        elif system == 'Darwin':
            lib_proxy = os.path.join(os.getcwd(), "libproxy5.dylib")
        else:
            raise OSError("Unsupported operating system")
        print(f"【黑龙江社保】lib_proxy：{lib_proxy}")
        if os.path.exists(lib_proxy):
            print(f"【黑龙江社保】lib_proxy存在")
        else:
            raise OSError("【黑龙江社保】lib_proxy：文件不存在！")
        self.lib = ctypes.cdll.LoadLibrary(lib_proxy)
        # 定义函数参数/返回类型
        self.lib.PostUrlWithProxy.argtypes = [
            ctypes.c_char_p,  # method
            ctypes.c_char_p,  # url
            ctypes.c_char_p,  # headers
            ctypes.c_char_p,  # proxy
            ctypes.c_char_p,  # allow_redirects
            ctypes.c_char_p   # post body
        ]
        self.lib.PostUrlWithProxy.restype = ctypes.c_char_p

    def call_go_proxy(self, method, get_url, headers, proxy_url, post_body, disable_redirect):
        """
        通过dll文件提供外部调用
        Args:
            method: 请求方式, dll暂只支持post与get
            get_url: url地址
            headers: headers(dict)
            proxy_url: 代理ip()
            post_body: 请求数据(dict)
            disable_redirect: 禁止重定向

        Returns:

        """
        disable_redirect_str = "true" if disable_redirect else "false"
        # headers字典转换json
        headers_json = json.dumps(headers).encode("utf-8")
        # 调用 C 函数
        if isinstance(post_body, str):
            body = post_body.encode('utf-8')
        else:
            body = json.dumps(post_body).encode('utf-8')
        result = self.lib.PostUrlWithProxy(
            method.upper().encode('utf-8'),
            get_url.encode('utf-8'),
            headers_json,
            proxy_url.encode('utf-8'),
            disable_redirect_str.encode('utf-8'),
            body
        )
        # 解析 JSON 结果
        res_json = json.loads(result.decode('utf-8'))
        if res_json["error"]:
            raise Exception(res_json["error"])
        return res_json["result"]

    def dll_get_response(self, method, url, headers, proxy_url, post_body="", disable_redirect=True):
        try:
            result = self.call_go_proxy(method, url, headers, proxy_url, post_body, disable_redirect)
            if result["status_code"] not in [200, 302]:
                raise Exception(result["status"])
            return result
        except Exception as e:
            raise Exception(e)

if __name__ == '__main__':
    li = LibProxy(None)
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
    }
    res = li.dll_get_response("GET", "https://www.baidu.com", headers, "")
    print(res.get('cookies'))
