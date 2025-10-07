package com.chiyu.hduchat.common.properties;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;

import java.util.ArrayList;
import java.util.List;

/**
 * @author chiyu
 * @since 2025/10
 */
@Data
@ConfigurationProperties("hduchat.auth")
public class AuthProps {

    /**
     * 默认忽略拦截的URL集合
     */
    private List<String> skipUrl = new ArrayList();

    /**
     * salt
     */
    private String saltKey; //= "hduchat-salt";
}
