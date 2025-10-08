package com.chiyu.hduchat.component.auth.model;

import lombok.Data;
import lombok.experimental.Accessors;

/**
 * @author chiyu
 * @since 2025/10
 */
@Data
@Accessors(chain = true)
public class TokenInfo<T> {

    /**
     * Token
     */
    private String token;

    /**
     * 过期时间
     */
    private Long expiration;

    /**
     * 用户信息
     */
    private T user;
}
