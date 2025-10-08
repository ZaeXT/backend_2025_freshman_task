package com.chiyu.hduchat.common.utils;

import cn.hutool.json.JSONUtil;
import com.chiyu.hduchat.common.constant.CommonConst;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.SneakyThrows;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;

/**
 * 封装了与 HttpServletRequest、HttpServletResponse 相关的常用操作，旨在简化 Web 层请求 / 响应处理、信息提取等逻辑
 * @author chiyu
 * @since 2025/10
 */
public class ServletUtil {

    @SneakyThrows
    public static void write(HttpServletResponse response, R data) {
        response.setStatus(data.getCode());
        response.setHeader("Content-type", "application/json;charset=" + CommonConst.UTF_8);
        response.setCharacterEncoding(CommonConst.UTF_8);
        response.getWriter().write(JSONUtil.toJsonStr(data));
    }

    @SneakyThrows
    public static void write(HttpServletResponse response, int status, R data) {
        response.setStatus(status);
        response.setHeader("Content-type", "application/json;charset=" + CommonConst.UTF_8);
        response.setCharacterEncoding(CommonConst.UTF_8);
        response.getWriter().write(JSONUtil.toJsonStr(data));
    }

    public static HttpServletRequest getRequest() {
        ServletRequestAttributes servletRequestAttributes = (ServletRequestAttributes) RequestContextHolder.getRequestAttributes();
        if (servletRequestAttributes != null) {
            return servletRequestAttributes.getRequest();
        }
        return null;
    }

    public static String getAuthorizationToken() {
        String token = getRequest().getHeader("Authorization");
        if (token != null && token.toLowerCase().startsWith("bearer")) {
            return token.toLowerCase().replace("bearer", "").trim();
        }
        return null;
    }

    public static String getToken(String token) {
        if (token != null && token.toLowerCase().startsWith("bearer")) {
            return token.replace("bearer", "").trim();
        }
        return token;
    }

    public static String getIpAddr() {
        HttpServletRequest request = getRequest();
        if (request == null) {
            return "unknown";
        } else {
            String ip = request.getHeader("x-forwarded-for");
            if (ip == null || ip.isEmpty() || "unknown".equalsIgnoreCase(ip)) {
                ip = request.getHeader("Proxy-Client-IP");
            }

            if (ip == null || ip.isEmpty() || "unknown".equalsIgnoreCase(ip)) {
                ip = request.getHeader("X-Forwarded-For");
            }

            if (ip == null || ip.isEmpty() || "unknown".equalsIgnoreCase(ip)) {
                ip = request.getHeader("WL-Proxy-Client-IP");
            }

            if (ip == null || ip.isEmpty() || "unknown".equalsIgnoreCase(ip)) {
                ip = request.getHeader("X-Real-IP");
            }

            if (ip == null || ip.isEmpty() || "unknown".equalsIgnoreCase(ip)) {
                ip = request.getRemoteAddr();
            }

            if ("0:0:0:0:0:0:0:1".equals(ip)) {
                ip = "127.0.0.1";
            }

            if (ip.contains(",")) {
                ip = ip.split(",")[0];
            }
            return ip;
        }
    }
}
