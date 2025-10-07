package com.chiyu.hduchat.component.aicore.model.dto;

import lombok.Data;
import lombok.experimental.Accessors;

/**
 * @author chiyu
 * @since 2025/10
 */
@Data
@Accessors(chain = true)
public class EmbeddingR {

    /**
     * 写入到vector store的ID
     */
    private String vectorId;

    /**
     * 文档ID
     */
    private String docsId;

    /**
     * 知识库ID
     */
    private String knowledgeId;

    /**
     * Embedding后切片的文本
     */
    private String text;
}
