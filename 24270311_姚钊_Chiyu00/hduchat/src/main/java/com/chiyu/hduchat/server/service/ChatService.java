package com.chiyu.hduchat.server.service;

import com.chiyu.hduchat.component.aicore.model.dto.ChatReq;

/**
 * @author chiyu
 * @since 2025/10
 */
public interface ChatService {

    void chat(ChatReq req);


    /**
     * 文本请求
     */
    String text(ChatReq req);

}
