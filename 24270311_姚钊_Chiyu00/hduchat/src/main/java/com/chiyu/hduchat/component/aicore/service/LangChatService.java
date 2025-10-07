package com.chiyu.hduchat.component.aicore.service;

import com.chiyu.hduchat.component.aicore.model.dto.ChatReq;
import dev.langchain4j.service.TokenStream;

/**
 * @author chiyu
 * @since 2025/10
 */
public interface LangChatService {

    TokenStream chat(ChatReq req);

    String text(ChatReq req);

}
