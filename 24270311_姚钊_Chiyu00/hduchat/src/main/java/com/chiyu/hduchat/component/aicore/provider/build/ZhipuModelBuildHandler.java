package com.chiyu.hduchat.component.aicore.provider.build;

import com.chiyu.hduchat.aigc.model.entity.AigcModel;
import com.chiyu.hduchat.component.aicore.model.consts.ProviderEnum;
import com.chiyu.hduchat.component.aicore.model.consts.ChatErrorEnum;
import com.chiyu.hduchat.common.exception.ServiceException;
import dev.langchain4j.community.model.zhipu.ZhipuAiChatModel;
import dev.langchain4j.community.model.zhipu.ZhipuAiEmbeddingModel;
import dev.langchain4j.community.model.zhipu.ZhipuAiImageModel;
import dev.langchain4j.community.model.zhipu.ZhipuAiStreamingChatModel;
import dev.langchain4j.model.chat.ChatLanguageModel;
import dev.langchain4j.model.chat.StreamingChatLanguageModel;
import dev.langchain4j.model.embedding.EmbeddingModel;
import dev.langchain4j.model.image.ImageModel;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.StringUtils;
import org.springframework.stereotype.Component;

import java.time.Duration;

/**
 * @author chiyu
 * @since 2024-08-19
 */
@Slf4j
@Component
@AllArgsConstructor
public class ZhipuModelBuildHandler implements ModelBuildHandler {


    @Override
    public boolean whetherCurrentModel(AigcModel model) {
        return ProviderEnum.ZHIPU.name().equals(model.getProvider());
    }

    @Override
    public boolean basicCheck(AigcModel model) {
        if (StringUtils.isBlank(model.getApiKey())) {
            throw new ServiceException(ChatErrorEnum.API_KEY_IS_NULL.getErrorCode(),
                    ChatErrorEnum.API_KEY_IS_NULL.getErrorDesc(ProviderEnum.ZHIPU.name(), model.getType()));
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
            return ZhipuAiStreamingChatModel
                    .builder()
                    .apiKey(model.getApiKey())
                    .baseUrl(model.getBaseUrl())
                    .model(model.getModel())
                    .maxToken(model.getResponseLimit())
                    .temperature(model.getTemperature())
                    .topP(model.getTopP())
                    .logRequests(true)
                    .logResponses(true)
                    .callTimeout(Duration.ofMinutes(10))
                    .connectTimeout(Duration.ofMinutes(10))
                    .writeTimeout(Duration.ofMinutes(10))
                    .readTimeout(Duration.ofMinutes(10))
                    .build();
        } catch (ServiceException e) {
            log.error(e.getMessage());
            throw e;
        } catch (Exception e) {
            log.error("zhipu streaming chat 配置报错", e);
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
            return ZhipuAiChatModel
                    .builder()
                    .apiKey(model.getApiKey())
                    .baseUrl(model.getBaseUrl())
                    .model(model.getModel())
                    .maxToken(model.getResponseLimit())
                    .temperature(model.getTemperature())
                    .topP(model.getTopP())
                    .logRequests(true)
                    .logResponses(true)
                    .callTimeout(Duration.ofMinutes(10))
                    .connectTimeout(Duration.ofMinutes(10))
                    .writeTimeout(Duration.ofMinutes(10))
                    .readTimeout(Duration.ofMinutes(10))
                    .build();
        } catch (ServiceException e) {
            log.error(e.getMessage());
            throw e;
        } catch (Exception e) {
            log.error("zhipu chat 配置报错", e);
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
            return ZhipuAiEmbeddingModel
                    .builder()
                    .apiKey(model.getApiKey())
                    .model(model.getModel())
                    .baseUrl(model.getBaseUrl())
                    .logRequests(true)
                    .logResponses(true)
                    .callTimeout(Duration.ofMinutes(10))
                    .connectTimeout(Duration.ofMinutes(10))
                    .writeTimeout(Duration.ofMinutes(10))
                    .readTimeout(Duration.ofMinutes(10))
                    .dimensions(1024)
                    .build();
        } catch (ServiceException e) {
            log.error(e.getMessage());
            throw e;
        } catch (Exception e) {
            log.error("zhipu embedding 配置报错", e);
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
            return ZhipuAiImageModel
                    .builder()
                    .apiKey(model.getApiKey())
                    .model(model.getModel())
                    .baseUrl(model.getBaseUrl())
                    .logRequests(true)
                    .logResponses(true)
                    .callTimeout(Duration.ofMinutes(10))
                    .connectTimeout(Duration.ofMinutes(10))
                    .writeTimeout(Duration.ofMinutes(10))
                    .readTimeout(Duration.ofMinutes(10))
                    .build();
        } catch (ServiceException e) {
            log.error(e.getMessage());
            throw e;
        } catch (Exception e) {
            log.error("zhipu image 配置报错", e);
            return null;
        }
    }
}
