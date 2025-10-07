package com.chiyu.hduchat.upms.model.dto;

import com.chiyu.hduchat.upms.model.entity.SysDept;
import com.chiyu.hduchat.upms.model.entity.SysRole;
import com.chiyu.hduchat.upms.model.entity.SysUser;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.List;
import java.util.Set;

/**
 * 自定义Oauth2 授权成功后存储的用户数据
 *
 * @author chiyu
 * @since 2025/10
 */
@Data
@Accessors(chain = true)
public class UserInfo extends SysUser implements Serializable {
    private static final long serialVersionUID = 547891924677981054L;

    /**
     * 用户所属部门
     */
    private SysDept dept;

    /**
     * 部门名称
     */
    private String deptName;

    /**
     * 角色ID列表
     */
    private List<String> roleIds;

    /**
     * 用户角色列表
     */
    private List<SysRole> roles;

    /**
     * 用户权限标识
     */
    private Set<String> perms;
}
