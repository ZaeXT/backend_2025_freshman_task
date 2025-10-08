package com.chiyu.hduchat.configuration;

import com.chiyu.hduchat.common.properties.AuthProps;
import com.chiyu.hduchat.common.properties.ChatProps;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Configuration;

/**
 * @author chiyu
 * @since 2025/10
 */
@Configuration
@EnableConfigurationProperties({
        AuthProps.class,
        ChatProps.class
})
public class CommonAutoConfiguration {

}
