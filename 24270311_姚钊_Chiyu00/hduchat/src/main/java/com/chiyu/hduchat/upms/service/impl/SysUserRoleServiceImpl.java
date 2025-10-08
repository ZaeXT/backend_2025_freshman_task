package com.chiyu.hduchat.upms.service.impl;

import com.chiyu.hduchat.upms.model.entity.SysRole;
import com.chiyu.hduchat.upms.model.entity.SysUser;
import com.chiyu.hduchat.upms.model.entity.SysUserRole;
import com.chiyu.hduchat.upms.mapper.SysUserRoleMapper;
import com.chiyu.hduchat.upms.service.SysUserRoleService;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

/**
 * 用户角色关联表(UserRole)表服务实现类
 *
 * @author chiyu
 * @since 2025/10
 */
@Service
public class SysUserRoleServiceImpl extends ServiceImpl<SysUserRoleMapper, SysUserRole> implements SysUserRoleService {

    @Override
    public List<SysUser> getUserListByRoleId(String roleId) {
        return baseMapper.getUserListByRoleId(roleId);
    }

    @Override
    public List<SysRole> getRoleListByUserId(String userId) {
        return baseMapper.getRoleListByUserId(userId);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteUserRolesByUserId(String userId) {
        baseMapper.delete(new LambdaQueryWrapper<SysUserRole>().eq(SysUserRole::getUserId, userId));
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteUserRolesByRoleId(String roleId) {
        baseMapper.delete(new LambdaQueryWrapper<SysUserRole>().eq(SysUserRole::getRoleId, roleId));
    }
}
