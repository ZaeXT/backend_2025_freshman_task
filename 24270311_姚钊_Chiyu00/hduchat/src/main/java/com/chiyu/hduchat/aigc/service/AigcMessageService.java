package com.chiyu.hduchat.aigc.service;

import com.chiyu.hduchat.aigc.model.entity.AigcConversation;
import com.chiyu.hduchat.aigc.model.entity.AigcMessage;
import com.chiyu.hduchat.common.utils.QueryPage;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.service.IService;

import java.util.List;

/**
 * @author chiyu
 * @since 2025/10
 */
public interface AigcMessageService extends IService<AigcMessage> {

    /**
     * 获取会话列表
     */
    List<AigcConversation> conversations(String userId);

    /**
     * 获取会话分页列表
     */
    IPage<AigcConversation> conversationPages(AigcConversation data, QueryPage queryPage);

    /**
     * 新增会话
     */
    AigcConversation addConversation(AigcConversation conversation);

    /**
     * 修改会话
     */
    void updateConversation(AigcConversation conversation);

    /**
     * 删除会话
     */
    void delConversation(String conversationId);

    AigcMessage addMessage(AigcMessage message);

    void clearMessage(String conversationId);

    List<AigcMessage> getMessages(String conversationId);

    List<AigcMessage> getMessages(String conversationId, String userId);
}

