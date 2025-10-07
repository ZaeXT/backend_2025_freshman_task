package com.chiyu.hduchat.common.constant;

import lombok.Getter;

/**
 * @author chiyu
 * @since 2025/10
 */
@Getter
public enum RoleEnum {
    USER("user"),
    ASSISTANT("assistant"),
    SYSTEM("system"),
    ;

    private final String name;

    RoleEnum(String name) {
        this.name = name;
    }
}
