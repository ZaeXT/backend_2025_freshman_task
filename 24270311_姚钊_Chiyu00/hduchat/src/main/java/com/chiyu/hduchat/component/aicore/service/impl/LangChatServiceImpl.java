package com.chiyu.hduchat.component.aicore.service.impl;

import cn.hutool.core.util.IdUtil;
import cn.hutool.core.util.StrUtil;
import com.chiyu.hduchat.common.properties.ChatProps;
import com.chiyu.hduchat.component.aicore.model.dto.ChatReq;
import com.chiyu.hduchat.component.aicore.provider.ModelProvider;
import com.chiyu.hduchat.component.aicore.service.Agent;
import com.chiyu.hduchat.component.aicore.service.LangChatService;
import dev.langchain4j.memory.chat.MessageWindowChatMemory;
import dev.langchain4j.model.chat.ChatLanguageModel;
import dev.langchain4j.model.chat.StreamingChatLanguageModel;
import dev.langchain4j.service.AiServices;
import dev.langchain4j.service.TokenStream;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

/**
 * @author chiyu
 * @since 2025/10
 */
@Slf4j
@Service
@AllArgsConstructor
public class LangChatServiceImpl implements LangChatService {

    private final ModelProvider provider;
    private final ChatProps chatProps;

    private AiServices<Agent> build(StreamingChatLanguageModel streamModel, ChatLanguageModel model, ChatReq req) {
        AiServices<Agent> aiServices = AiServices.builder(Agent.class)
                .chatMemoryProvider(memoryId -> MessageWindowChatMemory.builder()
                        .id(req.getConversationId())
                        .chatMemoryStore(new PersistentChatMemoryStore())
                        .maxMessages(chatProps.getMemoryMaxMessage())
                        .build());
        if (StrUtil.isNotBlank(req.getPromptText())) {
            aiServices.systemMessageProvider(memoryId -> req.getPromptText());
        }
        if (streamModel != null) {
            aiServices.streamingChatLanguageModel(streamModel);
        }
        if (model != null) {
            aiServices.chatLanguageModel(model);
        }
        return aiServices;
    }

    @Override
    public TokenStream chat(ChatReq req) {
        StreamingChatLanguageModel model = provider.stream(req.getModelId());
        if (StrUtil.isBlank(req.getConversationId())) {
            req.setConversationId(IdUtil.simpleUUID());
        }

        AiServices<Agent> aiServices = build(model, null, req);

        /*if (StrUtil.isNotBlank(req.getKnowledgeId())) {
            req.getKnowledgeIds().add(req.getKnowledgeId());
        }

        if (req.getKnowledgeIds() != null && !req.getKnowledgeIds().isEmpty()) {
            // TODO 知识库
        }*/
        Agent agent = aiServices.build();
        return agent.stream(req.getConversationId(), req.getMessage());
    }


    @Override
    public String text(ChatReq req) {
        if (StrUtil.isBlank(req.getConversationId())) {
            req.setConversationId(IdUtil.simpleUUID());
        }

        try {
            ChatLanguageModel model = provider.text(req.getModelId());
            Agent agent = build(null, model, req).build();
            String text = agent.text(req.getConversationId(), req.getMessage());
            return text;
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }

}
