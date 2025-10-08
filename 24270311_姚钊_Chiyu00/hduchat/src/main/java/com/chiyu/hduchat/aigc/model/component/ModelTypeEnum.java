package com.chiyu.hduchat.aigc.model.component;

import lombok.Getter;

/**
 * @author chiyu
 * @since 2025/10
 */
@Getter
public enum ModelTypeEnum {

    CHAT,
    EMBEDDING,
    TEXT_IMAGE,
    WEB_SEARCH;
}
