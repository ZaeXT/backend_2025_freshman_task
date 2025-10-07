package com.chiyu.hduchat.upms.mapper;

import com.chiyu.hduchat.upms.model.entity.SysRole;
import com.chiyu.hduchat.upms.model.entity.SysUser;
import com.chiyu.hduchat.upms.model.entity.SysUserRole;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Mapper;

import java.util.List;

/**
 * 用户角色关联表(UserRole)表数据库访问层
 *
 * @author chiyu
 * @since 2025/10
 */
@Mapper
public interface SysUserRoleMapper extends BaseMapper<SysUserRole> {

    /**
     * 根据RoleID查询User集合
     */
    List<SysUser> getUserListByRoleId(String roleId);

    /**
     * 根据UserID查询Role集合
     */
    List<SysRole> getRoleListByUserId(String userId);
}
