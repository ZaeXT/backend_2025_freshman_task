package com.chiyu.hduchat.component.auth.interceptor;

import cn.dev33.satoken.interceptor.SaInterceptor;
import lombok.AllArgsConstructor;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

/**
 * @author chiyu
 * @since 2025/10
 */
@Component
@AllArgsConstructor
public class AuthInterceptor implements WebMvcConfigurer {

    private final StringRedisTemplate redisTemplate;

    @Override
    public void addInterceptors(InterceptorRegistry registry) {
        registry.addInterceptor(new SaInterceptor())
                .addPathPatterns("/**")
                .excludePathPatterns(
                        // Knife4j 前端页面
                        "/doc.html",
                        // Swagger 资源路径
                        "/swagger-resources/**",
                        "/v3/api-docs/**",
                        "/v3/api-docs.yaml",
                        // 静态资源
                        "/webjars/**",
                        "/favicon.ico",
                        // 已有的跳过路径
                        "/auth/login",
                        "/auth/logout",
                        "/auth/register",
                        "/auth/info"
                );
    }
}
