package com.chiyu.hduchat.aigc.model.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.baomidou.mybatisplus.extension.handlers.FastjsonTypeHandler;
import lombok.Data;
import lombok.experimental.Accessors;
import org.apache.ibatis.type.JdbcType;

import java.io.Serializable;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;

/**
 * @author chiyu
 * @since 2025/10
 */
@Data
@TableName(autoResultMap = true)
@Accessors(chain = true)
public class AigcApp implements Serializable {
    private static final long serialVersionUID = -94917153262781949L;

    /**
     * 主键
     */
    @TableId(type = IdType.ASSIGN_UUID)
    private String id;
    private String modelId;

    @TableField(typeHandler = FastjsonTypeHandler.class, jdbcType = JdbcType.VARCHAR)
    private List<String> knowledgeIds;

    /**
     * 名称
     */
    private String name;
    private String cover;

    /**
     * Prompt
     */
    private String prompt;

    /**
     * 应用描述
     */
    private String des;

    /**
     * 创建时间
     */
    private Date saveTime;
    private Date createTime;

    @TableField(exist = false)
    private AigcModel model;
    @TableField(exist = false)
    private List<AigcKnowledge> knowledges = new ArrayList<>();
}
