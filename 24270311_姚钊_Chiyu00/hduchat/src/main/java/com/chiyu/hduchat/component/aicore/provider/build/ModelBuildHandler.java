package com.chiyu.hduchat.component.aicore.provider.build;

import com.chiyu.hduchat.aigc.model.entity.AigcModel;
import dev.langchain4j.model.chat.ChatLanguageModel;
import dev.langchain4j.model.chat.StreamingChatLanguageModel;
import dev.langchain4j.model.embedding.EmbeddingModel;
import dev.langchain4j.model.image.ImageModel;

/**
 * @author chiyu
 * @since 2024-08-18 09:57
 */
public interface ModelBuildHandler {

    /**
     * 判断是不是当前模型
     */
    boolean whetherCurrentModel(AigcModel model);

    /**
     * basic check
     */
    boolean basicCheck(AigcModel model);

    /**
     * streaming chat build
     */
    StreamingChatLanguageModel buildStreamingChat(AigcModel model);

    /**
     * chat build
     */
    ChatLanguageModel buildChatLanguageModel(AigcModel model);

    /**
     * embedding config
     */
    EmbeddingModel buildEmbedding(AigcModel model);

    /**
     * image config
     */
    ImageModel buildImage(AigcModel model);

}
