package com.chiyu.hduchat.component.aicore.service;

import dev.langchain4j.service.MemoryId;
import dev.langchain4j.service.TokenStream;
import dev.langchain4j.service.UserMessage;

/**
 * @author chiyu
 * @since 2025/10
 */
public interface Agent {

    TokenStream stream(@MemoryId String id, @UserMessage String message);

    String text(@MemoryId String id, @UserMessage String message);
}
