package com.chiyu.hduchat.server.controller;

import cn.dev33.satoken.annotation.SaCheckPermission;
import cn.hutool.core.util.StrUtil;
import com.chiyu.hduchat.aigc.model.entity.AigcApp;
import com.chiyu.hduchat.aigc.model.entity.AigcMessage;
import com.chiyu.hduchat.aigc.model.entity.AigcModel;
import com.chiyu.hduchat.aigc.service.AigcAppService;
import com.chiyu.hduchat.aigc.service.AigcMessageService;
import com.chiyu.hduchat.aigc.service.AigcModelService;
import com.chiyu.hduchat.component.aicore.service.impl.PersistentChatMemoryStore;
import com.chiyu.hduchat.component.aicore.model.dto.ChatReq;
import com.chiyu.hduchat.component.aicore.model.dto.ChatRes;
import com.chiyu.hduchat.component.aicore.model.consts.PromptConst;
import com.chiyu.hduchat.common.properties.ChatProps;
import com.chiyu.hduchat.component.aicore.utils.PromptUtil;
import com.chiyu.hduchat.component.aicore.utils.StreamEmitter;
import com.chiyu.hduchat.common.constant.RoleEnum;
import com.chiyu.hduchat.common.utils.R;
import com.chiyu.hduchat.server.service.ChatService;
import com.chiyu.hduchat.component.auth.AuthUtil;
import dev.langchain4j.data.message.AiMessage;
import dev.langchain4j.data.message.ChatMessage;
import dev.langchain4j.data.message.SystemMessage;
import dev.langchain4j.data.message.UserMessage;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * @author chiyu
 * @since 2025/10
 */
@Slf4j
@RequestMapping("/aigc")
@RestController
@AllArgsConstructor
public class ChatController {

    private final ChatService chatService;
    private final AigcMessageService messageService;
    private final AigcModelService aigcModelService;
    private final AigcAppService appService;
    private final ChatProps chatProps;

    @PostMapping("/chat/completions")
    @SaCheckPermission("chat:completions")
    public SseEmitter chat(@RequestBody ChatReq req) {
        StreamEmitter emitter = new StreamEmitter();
        req.setEmitter(emitter);
        req.setUserId(AuthUtil.getUserId());
        req.setUsername(AuthUtil.getUsername());
        ExecutorService executor = Executors.newSingleThreadExecutor();
        req.setExecutor(executor);
        return emitter.streaming(executor, () -> {
            chatService.chat(req);
        });
    }

    @GetMapping("/app/info")
    public R<AigcApp> appInfo(@RequestParam String appId, String conversationId) {
        AigcApp app = appService.getById(appId);
        if (StrUtil.isBlank(conversationId)) {
            conversationId = app.getId();
        }

        if (StrUtil.isNotBlank(app.getPrompt())) {
            // initialize chat memory
            SystemMessage message = new SystemMessage(app.getPrompt());
            PersistentChatMemoryStore.init(conversationId, message);
        }

        return R.ok(app);
    }

    @GetMapping("/chat/messages/{conversationId}")
    public R messages(@PathVariable String conversationId) {
        List<AigcMessage> list = messageService.getMessages(conversationId, String.valueOf(AuthUtil.getUserId()));

        // initialize chat memory
        List<ChatMessage> chatMessages = new ArrayList<>();
        list.forEach(item -> {
            if (chatMessages.size() >= chatProps.getMemoryMaxMessage()) {
                return;
            }
            if (item.getRole().equals(RoleEnum.ASSISTANT.getName())) {
                chatMessages.add(new AiMessage(item.getMessage()));
            } else {
                chatMessages.add(new UserMessage(item.getMessage()));
            }
        });
        PersistentChatMemoryStore.init(conversationId, chatMessages);
        return R.ok(list);
    }

    @DeleteMapping("/chat/messages/clean/{conversationId}")
    @SaCheckPermission("chat:messages:clean")
    public R cleanMessage(@PathVariable String conversationId) {
        messageService.clearMessage(conversationId);

        // clean chat memory
        PersistentChatMemoryStore.clean(conversationId);
        return R.ok();
    }

    @PostMapping("/chat/mindmap")
    public R mindmap(@RequestBody ChatReq req) {
        req.setPrompt(PromptUtil.build(req.getMessage(), PromptConst.MINDMAP));
        return R.ok(new ChatRes(chatService.text(req)));
    }

    //todo 图片生成
    /*@PostMapping("/chat/image")
    public R image(@RequestBody ImageR req) {
        req.setPrompt(PromptUtil.build(req.getMessage(), PromptConst.IMAGE));
        return R.ok(chatService.image(req));
    }*/

    @GetMapping("/chat/getImageModels")
    public R<List<AigcModel>> getImageModels() {
        List<AigcModel> list = aigcModelService.getImageModels();
        list.forEach(i -> {
            i.setApiKey(null);
            i.setSecretKey(null);
        });
        return R.ok(list);
    }
}
