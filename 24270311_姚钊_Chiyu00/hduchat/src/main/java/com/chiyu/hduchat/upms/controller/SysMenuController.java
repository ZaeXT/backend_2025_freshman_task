package com.chiyu.hduchat.upms.controller;

import cn.dev33.satoken.annotation.SaCheckPermission;
import com.chiyu.hduchat.common.annotation.ApiLog;
import com.chiyu.hduchat.common.utils.R;
import com.chiyu.hduchat.upms.model.dto.MenuTree;
import com.chiyu.hduchat.upms.model.entity.SysMenu;
import com.chiyu.hduchat.upms.service.SysMenuService;
import com.chiyu.hduchat.component.auth.AuthUtil;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * 菜单表(Menu)表控制层
 *
 * @author chiyu
 * @since 2025/10
 */
@RestController
@RequiredArgsConstructor
@RequestMapping("/upms/menu")
public class SysMenuController {

    private final SysMenuService sysMenuService;

    @GetMapping("/tree")
    public R<List<MenuTree<SysMenu>>> tree(SysMenu sysMenu) {
        return R.ok(sysMenuService.tree(sysMenu));
    }

    @GetMapping("/build")
    public R<List<MenuTree<SysMenu>>> build() {
        return R.ok(sysMenuService.build(AuthUtil.getUserId()));
    }

    @GetMapping("/list")
    public R<List<SysMenu>> list(SysMenu sysMenu) {
        return R.ok(sysMenuService.list(sysMenu));
    }

    @GetMapping("/{id}")
    public R<SysMenu> findById(@PathVariable String id) {
        return R.ok(sysMenuService.getById(id));
    }

    @PostMapping
    @ApiLog("新增菜单")
    @SaCheckPermission("upms:menu:add")
    public R add(@RequestBody SysMenu sysMenu) {
        sysMenuService.add(sysMenu);
        return R.ok();
    }

    @PutMapping
    @ApiLog("修改菜单")
    @SaCheckPermission("upms:menu:update")
    public R update(@RequestBody SysMenu sysMenu) {
        sysMenuService.update(sysMenu);
        return R.ok();
    }

    @DeleteMapping("/{id}")
    @ApiLog("删除菜单")
    @SaCheckPermission("upms:menu:delete")
    public R delete(@PathVariable String id) {
        sysMenuService.delete(id);
        return R.ok();
    }
}
