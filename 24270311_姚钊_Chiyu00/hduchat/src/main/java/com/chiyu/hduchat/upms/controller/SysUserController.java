package com.chiyu.hduchat.upms.controller;

import cn.dev33.satoken.annotation.SaCheckPermission;
import cn.hutool.core.lang.Dict;
import com.chiyu.hduchat.common.annotation.ApiLog;
import com.chiyu.hduchat.common.exception.ServiceException;
import com.chiyu.hduchat.common.properties.AuthProps;
import com.chiyu.hduchat.common.utils.MybatisUtil;
import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.common.utils.R;
import com.chiyu.hduchat.upms.model.dto.UserInfo;
import com.chiyu.hduchat.upms.model.entity.SysUser;
import com.chiyu.hduchat.upms.service.SysUserService;
import com.chiyu.hduchat.component.auth.AuthUtil;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * 用户表(User)表控制层
 *
 * @author chiyu
 * @since 2025/10
 */
@RestController
@RequiredArgsConstructor
@RequestMapping("/upms/user")
public class SysUserController {

    private final SysUserService sysUserService;
    private final AuthProps authProps;

    @GetMapping("/checkName")
    public R<Boolean> checkName(UserInfo sysUser) {
        return R.ok(sysUserService.checkName(sysUser));
    }

    @GetMapping("/list")
    public R<List<SysUser>> list(SysUser sysUser) {
        return R.ok(sysUserService.list(sysUser));
    }

    @GetMapping("/page")
    public R<Dict> page(UserInfo user, QueryPage queryPage) {
        return R.ok(MybatisUtil.getData(sysUserService.page(user, queryPage)));
    }

    @GetMapping("/{id}")
    public R<UserInfo> findById(@PathVariable String id) {
        return R.ok(sysUserService.findById(id));
    }

    @PostMapping
    @ApiLog("新增用户")
    @SaCheckPermission("upms:user:add")
    public R<SysUser> add(@RequestBody UserInfo user) {
        user.setPassword(AuthUtil.encode(authProps.getSaltKey(), user.getPassword()));
        sysUserService.add(user);
        return R.ok();
    }

    @PutMapping
    @ApiLog("修改用户")
    @SaCheckPermission("upms:user:update")
    public R update(@RequestBody UserInfo user) {
        sysUserService.update(user);
        return R.ok();
    }

    @DeleteMapping("/{id}")
    @ApiLog("删除用户")
    @SaCheckPermission("upms:user:delete")
    public R delete(@PathVariable String id) {
        SysUser user = sysUserService.getById(id);
        if (user != null) {
            sysUserService.delete(id, user.getUsername());
        }
        return R.ok();
    }

    @PutMapping("/resetPass")
    @ApiLog("重置密码")
    @SaCheckPermission("upms:user:reset")
    public R resetPass(@RequestBody UserInfo data) {
        SysUser user = sysUserService.getById(data.getId());
        if (user != null) {
            sysUserService.reset(data.getId(), data.getPassword(), user.getUsername());
        }
        return R.ok();
    }

    @PutMapping("/updatePass")
    @ApiLog("修改密码")
    @SaCheckPermission("upms:user:updatePass")
    public R updatePass(@RequestBody UserInfo data) {
        SysUser user = sysUserService.getById(data.getId());
        if (user == null || !AuthUtil.decrypt(authProps.getSaltKey(), user.getPassword()).equals(data.getPassword())) {
            throw new ServiceException("Old password entered incorrectly, please re-enter");
        }
        user.setPassword(AuthUtil.encode(authProps.getSaltKey(), data.getPassword()));
        return R.ok();
    }
}
