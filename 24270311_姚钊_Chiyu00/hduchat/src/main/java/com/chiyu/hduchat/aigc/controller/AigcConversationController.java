package com.chiyu.hduchat.aigc.controller;

import cn.dev33.satoken.annotation.SaCheckPermission;
import com.chiyu.hduchat.aigc.model.entity.AigcConversation;
import com.chiyu.hduchat.aigc.model.entity.AigcMessage;
import com.chiyu.hduchat.aigc.service.AigcMessageService;
import com.chiyu.hduchat.common.annotation.ApiLog;
import com.chiyu.hduchat.common.utils.MybatisUtil;
import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.common.utils.R;
import com.chiyu.hduchat.common.utils.ServletUtil;
import com.chiyu.hduchat.component.auth.AuthUtil;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * @author chiyu
 * @since 2025/10
 */
@Slf4j
@RestController
@RequestMapping("/aigc/conversation")
@AllArgsConstructor
public class AigcConversationController {

    private final AigcMessageService aigcMessageService;

    /**
     * conversation list, filter by user
     */
    @GetMapping("/list")
    public R conversations() {
        return R.ok(aigcMessageService.conversations(String.valueOf(AuthUtil.getUserId())));
    }

    /**
     * conversation page
     */
    @GetMapping("/page")
    public R list(AigcConversation data, QueryPage queryPage) {
        return R.ok(MybatisUtil.getData(aigcMessageService.conversationPages(data, queryPage)));
    }

    @PostMapping
    @ApiLog("添加会话窗口")
    @SaCheckPermission("aigc:conversation:add")
    public R addConversation(@RequestBody AigcConversation conversation) {
        conversation.setUserId(String.valueOf(AuthUtil.getUserId()));
        return R.ok(aigcMessageService.addConversation(conversation));
    }

    @PutMapping
    @ApiLog("更新会话窗口")
    @SaCheckPermission("aigc:conversation:update")
    public R updateConversation(@RequestBody AigcConversation conversation) {
        if (conversation.getId() == null) {
            return R.fail("conversation id is null");
        }
        aigcMessageService.updateConversation(conversation);
        return R.ok();
    }

    @DeleteMapping("/{conversationId}")
    @ApiLog("删除会话窗口")
    @SaCheckPermission("aigc:conversation:delete")
    public R delConversation(@PathVariable String conversationId) {
        aigcMessageService.delConversation(conversationId);
        return R.ok();
    }

    @DeleteMapping("/message/{conversationId}")
    @ApiLog("清空会话窗口数据")
    @SaCheckPermission("aigc:conversation:clear")
    public R clearMessage(@PathVariable String conversationId) {
        aigcMessageService.clearMessage(conversationId);
        return R.ok();
    }

    /**
     * get messages with conversationId
     */
    @GetMapping("/messages/{conversationId}")
    public R getMessages(@PathVariable String conversationId) {
        List<AigcMessage> list = aigcMessageService.getMessages(conversationId);
        return R.ok(list);
    }

    /**
     * add message in conversation
     */
    @PostMapping("/message")
    public R addMessage(@RequestBody AigcMessage message) {
        message.setIp(ServletUtil.getIpAddr());
        return R.ok(aigcMessageService.addMessage(message));
    }
}
