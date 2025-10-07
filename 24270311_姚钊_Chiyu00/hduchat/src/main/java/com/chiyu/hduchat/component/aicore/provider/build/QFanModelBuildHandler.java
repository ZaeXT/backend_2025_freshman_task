package com.chiyu.hduchat.component.aicore.provider.build;

import com.chiyu.hduchat.aigc.model.entity.AigcModel;
import com.chiyu.hduchat.component.aicore.model.consts.ProviderEnum;
import com.chiyu.hduchat.component.aicore.model.consts.ChatErrorEnum;
import com.chiyu.hduchat.common.exception.ServiceException;
import dev.langchain4j.community.model.qianfan.QianfanChatModel;
import dev.langchain4j.community.model.qianfan.QianfanEmbeddingModel;
import dev.langchain4j.community.model.qianfan.QianfanStreamingChatModel;
import dev.langchain4j.model.chat.ChatLanguageModel;
import dev.langchain4j.model.chat.StreamingChatLanguageModel;
import dev.langchain4j.model.embedding.EmbeddingModel;
import dev.langchain4j.model.image.ImageModel;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.StringUtils;
import org.springframework.stereotype.Component;

/**
 * @author chiyu
 * @since 2024-08-19 10:08
 */
@Slf4j
@Component
public class QFanModelBuildHandler implements ModelBuildHandler {

    @Override
    public boolean whetherCurrentModel(AigcModel model) {
        return ProviderEnum.Q_FAN.name().equals(model.getProvider());
    }

    @Override
    public boolean basicCheck(AigcModel model) {
        if (StringUtils.isBlank(model.getApiKey())) {
            throw new ServiceException(ChatErrorEnum.API_KEY_IS_NULL.getErrorCode(),
                    ChatErrorEnum.API_KEY_IS_NULL.getErrorDesc(ProviderEnum.Q_FAN.name(), model.getType()));
        }
        if (StringUtils.isBlank(model.getSecretKey())) {
            throw new ServiceException(ChatErrorEnum.SECRET_KEY_IS_NULL.getErrorCode(),
                    ChatErrorEnum.SECRET_KEY_IS_NULL.getErrorDesc(ProviderEnum.Q_FAN.name(), model.getType()));
        }
        return true;
    }

    @Override
    public StreamingChatLanguageModel buildStreamingChat(AigcModel model) {
        try {
            if (!whetherCurrentModel(model)) {
                return null;
            }
            if (!basicCheck(model)) {
                return null;
            }
            return QianfanStreamingChatModel
                    .builder()
                    .apiKey(model.getApiKey())
                    .secretKey(model.getSecretKey())
                    .modelName(model.getModel())
                    .baseUrl(model.getBaseUrl())
                    .temperature(model.getTemperature())
                    .topP(model.getTopP())
                    .logRequests(true)
                    .logResponses(true)
                    .build();
        } catch (ServiceException e) {
            log.error(e.getMessage());
            throw e;
        } catch (Exception e) {
            log.error("Qianfan  streaming chat 配置报错", e);
            return null;
        }
    }

    @Override
    public ChatLanguageModel buildChatLanguageModel(AigcModel model) {
        try {
            if (!whetherCurrentModel(model)) {
                return null;
            }
            if (!basicCheck(model)) {
                return null;
            }
            return QianfanChatModel
                    .builder()
                    .apiKey(model.getApiKey())
                    .secretKey(model.getSecretKey())
                    .modelName(model.getModel())
                    .baseUrl(model.getBaseUrl())
                    .temperature(model.getTemperature())
                    .topP(model.getTopP())
                    .logRequests(true)
                    .logResponses(true)
                    .build();
        } catch (ServiceException e) {
            log.error(e.getMessage());
            throw e;
        } catch (Exception e) {
            log.error("Qianfan chat 配置报错", e);
            return null;
        }


    }

    @Override
    public EmbeddingModel buildEmbedding(AigcModel model) {
        try {
            if (!whetherCurrentModel(model)) {
                return null;
            }
            if (!basicCheck(model)) {
                return null;
            }
            return QianfanEmbeddingModel
                    .builder()
                    .apiKey(model.getApiKey())
                    .modelName(model.getModel())
                    .secretKey(model.getSecretKey())
                    .logRequests(true)
                    .logResponses(true)
                    .build();
        } catch (ServiceException e) {
            log.error(e.getMessage());
            throw e;
        } catch (Exception e) {
            log.error("Qianfan embedding 配置报错", e);
            return null;
        }
    }

    @Override
    public ImageModel buildImage(AigcModel model) {
        try {
            if (!whetherCurrentModel(model)) {
                return null;
            }
            if (!basicCheck(model)) {
                return null;
            }
            return null;
        } catch (ServiceException e) {
            log.error(e.getMessage());
            throw e;
        } catch (Exception e) {
            log.error("Qianfan image 配置报错", e);
            return null;
        }

    }
}
