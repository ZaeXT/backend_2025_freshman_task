package com.chiyu.hduchat.upms.controller;

import cn.dev33.satoken.annotation.SaCheckPermission;
import cn.hutool.core.lang.Dict;
import com.chiyu.hduchat.common.annotation.ApiLog;
import com.chiyu.hduchat.common.utils.MybatisUtil;
import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.common.utils.R;
import com.chiyu.hduchat.upms.model.dto.SysRoleDTO;
import com.chiyu.hduchat.upms.model.entity.SysRole;
import com.chiyu.hduchat.upms.service.SysRoleService;
import com.chiyu.hduchat.component.auth.AuthUtil;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * 角色表(Role)表控制层
 *
 * @author chiyu
 * @since 2025/10
 */
@RestController
@RequiredArgsConstructor
@RequestMapping("/upms/role")
public class SysRoleController {

    private final SysRoleService sysRoleService;

    @GetMapping("/list")
    public R<List<SysRole>> list(SysRole sysRole) {
        return R.ok(sysRoleService.list(new LambdaQueryWrapper<SysRole>()
                .ne(SysRole::getCode, AuthUtil.ADMINISTRATOR)));
    }

    @GetMapping("/page")
    public R<Dict> page(SysRole role, QueryPage queryPage) {
        return R.ok(MybatisUtil.getData(sysRoleService.page(role, queryPage)));
    }

    @GetMapping("/{id}")
    public R<SysRoleDTO> findById(@PathVariable String id) {
        return R.ok(sysRoleService.findById(id));
    }

    @PostMapping
    @ApiLog("新增角色")
    @SaCheckPermission("upms:role:add")
    public R add(@RequestBody SysRoleDTO sysRole) {
        sysRoleService.add(sysRole);
        return R.ok();
    }

    @PutMapping
    @ApiLog("修改角色")
    @SaCheckPermission("upms:role:update")
    public R update(@RequestBody SysRoleDTO sysRole) {
        sysRoleService.update(sysRole);
        return R.ok();
    }

    @DeleteMapping("/{id}")
    @ApiLog("删除角色")
    @SaCheckPermission("upms:role:delete")
    public R delete(@PathVariable String id) {
        sysRoleService.delete(id);
        return R.ok();
    }
}
