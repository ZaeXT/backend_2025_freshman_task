package com.chiyu.hduchat.common.properties;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;

/**
 * @author chiyu
 * @since 2025/10
 */
@Data
@ConfigurationProperties(prefix = "hduchat.chat")
public class ChatProps {

    /**
     * 上下文的长度
     */
    private Integer memoryMaxMessage = 20;

    /**
     * 前端渲染的消息长度（过长会导致页面渲染卡顿）
     */
    private Integer previewMaxMessage = 100;
}
