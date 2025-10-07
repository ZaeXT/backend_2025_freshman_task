package com.chiyu.hduchat.component.aicore.model.dto;

import lombok.Data;
import lombok.experimental.Accessors;

/**
 * @author chiyu
 * @since 2025/10
 */
@Data
@Accessors(chain = true)
public class ChatRes {

    private boolean isDone = false;

    private String message;

    private Integer usedToken;

    private long time;

    public ChatRes(String message) {
        this.message = message;
    }

    public ChatRes(Integer usedToken, long startTime) {
        this.isDone = true;
        this.usedToken = usedToken;
        this.time = System.currentTimeMillis() - startTime;
    }
}
