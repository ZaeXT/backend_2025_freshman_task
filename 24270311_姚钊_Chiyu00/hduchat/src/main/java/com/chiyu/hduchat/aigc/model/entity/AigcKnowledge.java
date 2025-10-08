package com.chiyu.hduchat.aigc.model.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.io.Serializable;
import java.util.ArrayList;
import java.util.List;

/**
 * @author chiyu
 * @since 2025/10
 */
@Data
@TableName(autoResultMap = true)
public class AigcKnowledge implements Serializable {
    private static final long serialVersionUID = 548724967827903685L;

    /**
     * 主键
     */
    @TableId(type = IdType.ASSIGN_UUID)
    private String id;

    private String embedStoreId;
    private String embedModelId;

    /**
     * 知识库名称
     */
    private String name;

    /**
     * 封面
     */
    private String cover;

    /**
     * 描述
     */
    private String des;

    /**
     * 创建时间
     */
    private String createTime;

    @TableField(exist = false)
    private Integer docsNum = 0;
    @TableField(exist = false)
    private Long totalSize = 0L;
    @TableField(exist = false)
    private List<AigcDocs> docs = new ArrayList<>();

    @TableField(exist = false)
    private AigcEmbedStore embedStore;
    @TableField(exist = false)
    private AigcModel embedModel;
}

