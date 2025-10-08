package com.chiyu.hduchat.aigc.model.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

/**
 * @author chiyu
 * @since 2025/10
 */
@Data
@Accessors(chain = true)
public class AigcAppApi implements Serializable {
    private static final long serialVersionUID = -94917153262781949L;

    /**
     * 主键
     */
    @TableId(type = IdType.ASSIGN_UUID)
    private String id;
    private String appId;
    private String apiKey;
    private String channel;
    private Date createTime;

    @TableField(exist = false)
    private AigcApp app;
}
