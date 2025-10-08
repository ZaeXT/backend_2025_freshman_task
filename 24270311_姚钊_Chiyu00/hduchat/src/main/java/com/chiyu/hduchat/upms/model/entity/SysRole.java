package com.chiyu.hduchat.upms.model.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;

/**
 * 角色表(Role)实体类
 *
 * @author chiyu
 * @since 2025/10
 */
@Data
@Accessors(chain = true)
public class SysRole implements Serializable {
    private static final long serialVersionUID = 547891924677981054L;

    /**
     * 主键
     */
    @TableId(type = IdType.ASSIGN_UUID)
    private String id;

    /**
     * 角色名称
     */
    private String name;

    /**
     * 角色别名
     */
    private String code;

    /**
     * 描述
     */
    private String des;
}
