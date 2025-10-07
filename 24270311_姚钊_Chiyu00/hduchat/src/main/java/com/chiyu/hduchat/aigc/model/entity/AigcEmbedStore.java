package com.chiyu.hduchat.aigc.model.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.io.Serializable;

/**
 * @author chiyu
 * @since 2025/10
 */
@Data
@TableName(autoResultMap = true)
public class AigcEmbedStore implements Serializable {
    private static final long serialVersionUID = 548724967827903685L;

    /**
     * 主键
     */
    @TableId(type = IdType.ASSIGN_UUID)
    private String id;
    private String name;

    private String provider;
    private String host;
    private Integer port;
    private String username;
    private String password;
    private String databaseName;
    private String tableName;
    private Integer dimension;
}
