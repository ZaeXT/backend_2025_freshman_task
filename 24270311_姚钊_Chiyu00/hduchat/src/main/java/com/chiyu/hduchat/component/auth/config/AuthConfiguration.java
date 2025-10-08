package com.chiyu.hduchat.component.auth.config;

import cn.dev33.satoken.context.SaHolder;
import cn.dev33.satoken.exception.NotPermissionException;
import cn.dev33.satoken.exception.NotRoleException;
import cn.dev33.satoken.filter.SaServletFilter;
import cn.dev33.satoken.router.SaRouter;
import cn.dev33.satoken.stp.StpUtil;
import cn.hutool.core.util.URLUtil;
import com.chiyu.hduchat.component.log.SysLogUtil;
import com.chiyu.hduchat.component.log.LogEvent;
import com.chiyu.hduchat.configuration.SpringContextHolder;
import com.chiyu.hduchat.common.properties.AuthProps;
import com.chiyu.hduchat.common.utils.R;
import com.chiyu.hduchat.upms.model.entity.SysLog;
import com.chiyu.hduchat.component.auth.AuthUtil;
import com.alibaba.fastjson.JSON;
import jakarta.servlet.http.HttpServletRequest;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.HttpStatus;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;

import java.util.Objects;

/**
 * @author chiyu
 * @since 2025/10
 */
@Slf4j
@Configuration
@AllArgsConstructor
public class AuthConfiguration {

    private final AuthProps authProps;
    private final String[] skipUrl = new String[]{
            "/auth/login",
            "/auth/logout",
            "/auth/register",
            "/auth/info",
    };

    @Bean
    public SaServletFilter saServletFilter() {
        return new SaServletFilter()
                .addInclude("/**")
                .addExclude("/favicon.ico")
                .addExclude(
                        "/doc.html",
                        "/webjars/**",
                        "/swagger-resources/**",
                        "/v3/api-docs/**")
                .setAuth(obj -> {
                    SaRouter
                            .match("/upms/**", "/aigc/**", "/app/**")
                            .check(StpUtil::checkLogin)
                            .notMatch(skipUrl)
                            .notMatch(authProps.getSkipUrl().toArray(new String[0]))
                    ;
                })
                .setError(this::handleError);
    }

    private String handleError(Throwable e) {
        if (e instanceof NotPermissionException || e instanceof NotRoleException) {
            String username = AuthUtil.getUsername();
            SysLog sysLog = SysLogUtil.build(SysLogUtil.TYPE_FAIL, HttpStatus.UNAUTHORIZED.getReasonPhrase(), null, null, username);
            SpringContextHolder.publishEvent(new LogEvent(sysLog));
        }

        HttpServletRequest request = ((ServletRequestAttributes) Objects.requireNonNull(RequestContextHolder.getRequestAttributes())).getRequest();
        log.error("Unauthorized request：{}", URLUtil.getPath(request.getRequestURI()));

        SaHolder.getResponse()
                .setStatus(HttpStatus.UNAUTHORIZED.value())
                .setHeader("Content-Type", "application/json;charset=UTF-8");
        return JSON.toJSONString(R.fail(HttpStatus.UNAUTHORIZED));
    }
}
