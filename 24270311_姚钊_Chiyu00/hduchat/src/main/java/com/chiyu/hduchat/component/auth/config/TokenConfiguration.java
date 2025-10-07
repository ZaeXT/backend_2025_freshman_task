package com.chiyu.hduchat.component.auth.config;

import cn.dev33.satoken.config.SaTokenConfig;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;

/**
 * @author chiyu
 * @since 2025/10
 */
@Configuration
public class TokenConfiguration {

    //todo 2. 生成唯一令牌（Token）
    @Bean
    @Primary
    public SaTokenConfig getTokenConfig() {
        return new SaTokenConfig()
                .setIsPrint(false)
                .setTokenName("Authorization")
                .setTimeout(24 * 60 * 60)
                .setTokenStyle("uuid")
                .setIsLog(false)
                .setIsReadCookie(false)
                ;//使用默认策略（允许多端登录）。
    }
}
