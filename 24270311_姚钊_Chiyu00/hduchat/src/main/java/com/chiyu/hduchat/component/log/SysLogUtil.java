package com.chiyu.hduchat.component.log;

import cn.hutool.core.util.URLUtil;
import cn.hutool.extra.servlet.JakartaServletUtil;
import cn.hutool.http.HttpUtil;
import com.chiyu.hduchat.configuration.SpringContextHolder;
import com.chiyu.hduchat.upms.model.entity.SysLog;
import jakarta.servlet.http.HttpServletRequest;
import lombok.SneakyThrows;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;

import java.util.Date;
import java.util.Objects;

/**
 * 构建Log实体类信息
 *
 * @author chiyu
 * @since 2025/10
 */
public class SysLogUtil {

    /**
     * 成功日志类型
     */
    public static final int TYPE_OK = 1;
    /**
     * 错误日志类型
     */
    public static final int TYPE_FAIL = 2;

    /**
     * 构建日志Log类信息
     *
     * @param type      日志类型
     * @param operation 日志描述
     * @param method    操作方法
     * @param time      耗时
     * @return Log类
     */
    @SneakyThrows
    public static SysLog build(Integer type, String operation, String method, Long time, String username) {
        HttpServletRequest request = ((ServletRequestAttributes)
                Objects.requireNonNull(RequestContextHolder.getRequestAttributes())).getRequest();

        return new SysLog()
                .setType(type)
                .setUsername(username)
                .setOperation(operation)
                .setCreateTime(new Date())
                .setIp(JakartaServletUtil.getClientIP(request))
                .setUrl(URLUtil.getPath(request.getRequestURI()))
                .setMethod(method)
                .setParams(HttpUtil.toParams(request.getParameterMap()))
                .setUserAgent(request.getHeader("user-agent"))
                .setTime(time);
    }

    /**
     * Spring事件发布：发布日志，写入到数据库
     *
     * @param type      日志类型
     * @param operation 描述
     */
    public static void publish(int type, String operation, String username) {
        SysLog sysLog = SysLogUtil.build(type, operation, null, null, username);
        SpringContextHolder.publishEvent(new LogEvent(sysLog));
    }
}
