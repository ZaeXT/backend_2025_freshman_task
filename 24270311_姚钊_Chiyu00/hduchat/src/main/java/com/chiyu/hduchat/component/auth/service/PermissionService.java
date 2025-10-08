package com.chiyu.hduchat.component.auth.service;

import cn.dev33.satoken.stp.StpInterface;
import com.chiyu.hduchat.component.auth.AuthUtil;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * @author chiyu
 * @since 2025/10
 */
@Slf4j
@Component
public class PermissionService implements StpInterface {
//作用是向框架提供当前登录用户的角色和权限信息，支撑系统的权限验证功能。
    @Override
    public List<String> getPermissionList(Object o, String s) {
        return AuthUtil.getPermissionNames();
    }

    @Override
    public List<String> getRoleList(Object o, String s) {
        return AuthUtil.getRoleNames();
    }
}
