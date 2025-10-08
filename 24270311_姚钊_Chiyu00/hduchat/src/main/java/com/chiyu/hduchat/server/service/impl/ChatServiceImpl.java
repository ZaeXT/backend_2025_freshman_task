package com.chiyu.hduchat.server.service.impl;

import cn.hutool.core.util.StrUtil;
import com.chiyu.hduchat.aigc.model.entity.AigcApp;
import com.chiyu.hduchat.aigc.model.entity.AigcMessage;
import com.chiyu.hduchat.aigc.service.AigcMessageService;
import com.chiyu.hduchat.component.aicore.service.LangChatService;
import com.chiyu.hduchat.component.aicore.model.dto.ChatReq;
import com.chiyu.hduchat.component.aicore.model.dto.ChatRes;
import com.chiyu.hduchat.component.aicore.utils.StreamEmitter;
import com.chiyu.hduchat.common.constant.RoleEnum;
import com.chiyu.hduchat.common.utils.ServletUtil;
import com.chiyu.hduchat.server.service.ChatService;
import com.chiyu.hduchat.component.store.AppStore;
import dev.langchain4j.model.output.TokenUsage;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.BeanUtils;
import org.springframework.stereotype.Service;

/**
 * @author chiyu
 * @since 2025/10
 */
@Slf4j
@Service
@AllArgsConstructor
public class ChatServiceImpl implements ChatService {

    private final LangChatService langChatService;
    private final AigcMessageService aigcMessageService;
    private final AppStore appStore;

    @Override
    public void chat(ChatReq req) {
        StreamEmitter emitter = req.getEmitter();
        long startTime = System.currentTimeMillis();
        StringBuilder text = new StringBuilder();

        if (StrUtil.isNotBlank(req.getAppId())) {
            AigcApp app = appStore.get(req.getAppId());
            if (app != null) {
                req.setModelId(app.getModelId());
                req.setPromptText(app.getPrompt());
                req.setKnowledgeIds(app.getKnowledgeIds());
            }
        }

        // save user message
        req.setRole(RoleEnum.USER.getName());
        saveMessage(req, 0, 0);

        try {
            langChatService
                    .chat(req)
                    .onNext(e -> {
                        text.append(e);
                        emitter.send(new ChatRes(e));
                    })
                    .onComplete((e) -> {
                        TokenUsage tokenUsage = e.tokenUsage();
                        ChatRes res = new ChatRes(tokenUsage.totalTokenCount(), startTime);
                        emitter.send(res);
                        emitter.complete();

                        // save assistant message
                        req.setMessage(text.toString());
                        req.setRole(RoleEnum.ASSISTANT.getName());
                        saveMessage(req, tokenUsage.inputTokenCount(), tokenUsage.outputTokenCount());
                    })
                    .onError((e) -> {
                        emitter.error(e.getMessage());
                        throw new RuntimeException(e.getMessage());
                    })
                    .start();
        } catch (Exception e) {
            e.printStackTrace();
            emitter.error(e.getMessage());
            throw new RuntimeException(e.getMessage());
        }
    }

    private void saveMessage(ChatReq req, Integer inputToken, Integer outputToken) {
        if (req.getConversationId() != null) {
            AigcMessage message = new AigcMessage();
            BeanUtils.copyProperties(req, message);
            message.setIp(ServletUtil.getIpAddr());
            message.setPromptTokens(inputToken);
            message.setTokens(outputToken);
            aigcMessageService.addMessage(message);
        }
    }

    @Override
    public String text(ChatReq req) {
        String text;
        try {
            text = langChatService.text(req);
        } catch (Exception e) {
            e.printStackTrace();
            throw new RuntimeException(e.getMessage());
        }
        return text;
    }

}
