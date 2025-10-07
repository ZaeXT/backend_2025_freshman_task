package com.chiyu.hduchat.upms.service.impl;

import cn.hutool.core.bean.BeanUtil;
import com.chiyu.hduchat.common.exception.ServiceException;
import com.chiyu.hduchat.common.utils.MybatisUtil;
import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.upms.model.dto.SysRoleDTO;
import com.chiyu.hduchat.upms.model.entity.SysRole;
import com.chiyu.hduchat.upms.model.entity.SysRoleMenu;
import com.chiyu.hduchat.upms.mapper.SysRoleMapper;
import com.chiyu.hduchat.upms.mapper.SysRoleMenuMapper;
import com.chiyu.hduchat.upms.mapper.SysUserRoleMapper;
import com.chiyu.hduchat.upms.service.SysRoleMenuService;
import com.chiyu.hduchat.upms.service.SysRoleService;
import com.chiyu.hduchat.upms.service.SysUserRoleService;
import com.chiyu.hduchat.component.auth.AuthUtil;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.core.toolkit.Wrappers;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import lombok.RequiredArgsConstructor;
import org.apache.commons.lang3.StringUtils;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

/**
 * 角色表(Role)表服务实现类
 *
 * @author chiyu
 * @since 2025/10
 */
@Service
@RequiredArgsConstructor
public class SysRoleServiceImpl extends ServiceImpl<SysRoleMapper, SysRole> implements SysRoleService {

    private final SysRoleMenuService sysRoleMenuService;
    private final SysUserRoleService sysUserRoleService;
    private final SysUserRoleMapper sysUserRoleMapper;
    private final SysRoleMenuMapper sysRoleMenuMapper;

    @Override
    public IPage<SysRole> page(SysRole role, QueryPage queryPage) {
        return baseMapper.selectPage(MybatisUtil.wrap(role, queryPage),
                Wrappers.<SysRole>lambdaQuery()
                        .like(StringUtils.isNotEmpty(role.getName()), SysRole::getName, role.getName())
        );
    }

    @Override
    public List<SysRole> findRolesByUserId(String id) {
        return sysUserRoleMapper.getRoleListByUserId(id);
    }

    private List<String> getMenuIdsByRoleId(String roleId) {
        List<SysRoleMenu> list = sysRoleMenuMapper.selectList(new LambdaQueryWrapper<SysRoleMenu>().eq(SysRoleMenu::getRoleId, roleId));
        return list.stream().map(SysRoleMenu::getMenuId).collect(Collectors.toList());
    }

    @Override
    public SysRoleDTO findById(String roleId) {
        SysRole role = this.getById(roleId);
        SysRoleDTO sysRole = BeanUtil.copyProperties(role, SysRoleDTO.class);
        sysRole.setMenuIds(getMenuIdsByRoleId(roleId));
        return sysRole;
    }

    public boolean checkCode(SysRoleDTO data) {
        if (AuthUtil.ADMINISTRATOR.equals(data.getCode()) || AuthUtil.DEFAULT_ROLE.equals(data.getCode())) {
            return false;
        }
        LambdaQueryWrapper<SysRole> queryWrapper = new LambdaQueryWrapper<SysRole>().eq(SysRole::getCode, data.getCode());
        if (data.getId() != null) {
            queryWrapper.ne(SysRole::getId, data.getId());
        }
        return baseMapper.selectList(queryWrapper).size() <= 0;
    }

    @Override
    public void add(SysRoleDTO sysRole) {
        if (!checkCode(sysRole)) {
            throw new ServiceException("该角色已存在");
        }
        this.save(sysRole);
        addMenus(sysRole);
    }

    @Override
    public void update(SysRoleDTO sysRole) {
        checkCode(sysRole);
        baseMapper.updateById(sysRole);
        addMenus(sysRole);
    }

    private void addMenus(SysRoleDTO sysRole) {
        List<String> menuIds = sysRole.getMenuIds();
        String id = sysRole.getId();
        if (menuIds != null) {
            // 先删除原有的关联
            sysRoleMenuService.deleteRoleMenusByRoleId(id);

            // 再新增关联
            List<SysRoleMenu> sysRoleMenuList = new ArrayList<>();
            menuIds.forEach(menuId -> sysRoleMenuList.add(new SysRoleMenu()
                    .setMenuId(menuId)
                    .setRoleId(id)));
            sysRoleMenuService.saveBatch(sysRoleMenuList);
        }
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void delete(String id) {
        SysRole sysRole = this.getById(id);
        if (!checkCode(BeanUtil.copyProperties(sysRole, SysRoleDTO.class))) {
            throw new ServiceException("该角色不可删除");
        }
        baseMapper.deleteById(id);
        sysRoleMenuService.deleteRoleMenusByRoleId(id);
        sysUserRoleService.deleteUserRolesByRoleId(id);
    }
}
