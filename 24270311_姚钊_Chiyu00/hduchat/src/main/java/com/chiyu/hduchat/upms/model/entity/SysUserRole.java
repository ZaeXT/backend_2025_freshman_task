package com.chiyu.hduchat.upms.model.entity;

import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;

/**
 * 用户角色关联表(UserRole)实体类
 *
 * @author chiyu
 * @since 2025/10
 */
@Data
@Accessors(chain = true)
public class SysUserRole implements Serializable {
    private static final long serialVersionUID = -24775118196826771L;

    /**
     * 用户ID
     */
    private String userId;

    /**
     * 角色ID
     */
    private String roleId;
}
