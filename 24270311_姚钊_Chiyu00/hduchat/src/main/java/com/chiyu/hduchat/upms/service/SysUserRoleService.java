package com.chiyu.hduchat.upms.service;

import com.chiyu.hduchat.upms.model.entity.SysRole;
import com.chiyu.hduchat.upms.model.entity.SysUser;
import com.chiyu.hduchat.upms.model.entity.SysUserRole;
import com.baomidou.mybatisplus.extension.service.IService;

import java.util.List;

/**
 * 用户角色关联表(UserRole)表服务接口
 *
 * @author chiyu
 * @since 2025/10
 */
public interface SysUserRoleService extends IService<SysUserRole> {

    /**
     * 根据RoleID查询User集合
     */
    List<SysUser> getUserListByRoleId(String roleId);

    /**
     * 根据UserID查询Role集合
     */
    List<SysRole> getRoleListByUserId(String userId);

    /**
     * 根据用户ID删除该用户的角色关联信息
     */
    void deleteUserRolesByUserId(String userId);

    /**
     * 根据角色ID删除该用户的角色关联信息
     */
    void deleteUserRolesByRoleId(String roleId);
}
