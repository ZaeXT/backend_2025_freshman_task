package com.chiyu.hduchat.upms.service;

import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.upms.model.dto.SysRoleDTO;
import com.chiyu.hduchat.upms.model.entity.SysRole;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.service.IService;

import java.util.List;

/**
 * 角色表(Role)表服务接口
 *
 * @author chiyu
 * @since 2025/10
 */
public interface SysRoleService extends IService<SysRole> {

    /**
     * 分页、条件查询
     */
    IPage<SysRole> page(SysRole role, QueryPage queryPage);

    /**
     * 根据用户ID查询其关联的所有角色
     */
    List<SysRole> findRolesByUserId(String id);

    /**
     * 根据ID查询
     */
    SysRoleDTO findById(String roleId);

    /**
     * 新增角色
     */
    void add(SysRoleDTO sysRole);

    /**
     * 修改角色
     */
    void update(SysRoleDTO sysRole);

    /**
     * 删除
     */
    void delete(String id);
}
