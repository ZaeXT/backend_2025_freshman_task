package com.chiyu.hduchat.upms.model.entity;

import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;

/**
 * 角色资源关联表(RoleMenu)实体类
 *
 * @author chiyu
 * @since 2025/10
 */
@Data
@Accessors(chain = true)
public class SysRoleMenu implements Serializable {
    private static final long serialVersionUID = 854296054659457976L;

    /**
     * 角色ID
     */
    private String roleId;

    /**
     * 菜单ID
     */
    private String menuId;
}
