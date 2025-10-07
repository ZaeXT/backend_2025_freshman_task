package com.chiyu.hduchat.component.auth;

import cn.dev33.satoken.secure.SaSecureUtil;
import cn.dev33.satoken.stp.StpUtil;
import com.chiyu.hduchat.common.constant.CacheConst;
import com.chiyu.hduchat.common.exception.AuthException;
import com.chiyu.hduchat.upms.model.dto.UserInfo;
import com.chiyu.hduchat.upms.model.entity.SysRole;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;

import java.util.ArrayList;
import java.util.List;
import java.util.Objects;

/**
 * 权限相关方法
 *
 * @author chiyu
 * @since 2025/10
 */
public class AuthUtil {

    /**
     * 系统预制固定超级管理员角色别名
     * 作用：提供一个角色摆脱权限体系的控制，允许词角色访问所有菜单权限
     * 使用：所有涉及根据角色查询的地方都排除对此角色的限制
     */
    public static final String ADMINISTRATOR = "administrator";
    public static final String DEMO_ROLE = "demo_env";
    public static final String DEFAULT_ROLE = "default_env";

    /**
     * 获取Request对象
     */
    public static HttpServletRequest getRequest() {
        return ((ServletRequestAttributes) Objects.requireNonNull(RequestContextHolder.getRequestAttributes())).getRequest();
    }

    /**
     * 获取Response对象
     */
    public static HttpServletResponse getResponse() {
        return ((ServletRequestAttributes) Objects.requireNonNull(RequestContextHolder.getRequestAttributes())).getResponse();
    }

    /**
     * 截取前端Token字符串中不包含`Bearer`的部分
     */
    public static String getToken(String token) {
        if (token != null && token.toLowerCase().startsWith("bearer")) {
            return token.replace("bearer", "").trim();
        }
        return token;
    }

    /**
     * 获取用户数据
     */
    //todo 7. 更新线程上下文（ThreadLocal 存储）
    public static UserInfo getUserInfo() {
        try {
            return (UserInfo) StpUtil.getSession().get(CacheConst.AUTH_USER_INFO_KEY);
        } catch (Exception e) {
            e.printStackTrace();
            throw new AuthException(403, "登录已失效，请重新登陆");
        }
    }
    public static boolean isDemoEnv() {
        try {
            UserInfo info =  (UserInfo) StpUtil.getSession().get(CacheConst.AUTH_USER_INFO_KEY);
            List<SysRole> roles = info.getRoles();
            if (roles != null && !roles.isEmpty()) {
                List<SysRole> list = roles.stream().filter(i -> i.getCode().equals(DEMO_ROLE)).toList();
                return !list.isEmpty();
            }
            return true;
        } catch (Exception ignored) {
            return true;
        }
    }

    /**
     * 获取用户名
     */
    public static String getUsername() {
        UserInfo userInfo = getUserInfo();
        if (userInfo == null) {
            return null;
        }
        return userInfo.getUsername();
    }

    /**
     * 获取登录用户ID
     */
    public static String getUserId() {
        UserInfo userInfo = getUserInfo();
        if (userInfo == null) {
            return null;
        }
        return userInfo.getId();
    }

    /**
     * 获取用户角色Id集合
     */
    public static List<String> getRoleIds() {
        /*
用户登录时：系统查询用户信息、关联的角色和权限，封装为UserInfo对象，存入 Sa-Token 会话（StpUtil.getSession()）。
         */
        UserInfo userInfo = getUserInfo();
        if (userInfo == null || userInfo.getRoleIds() == null) {
            return new ArrayList<>();
        }
        return userInfo.getRoleIds();
    }

    /**
     * 获取用户角色Alias集合
     */
    public static List<String> getRoleNames() {
        UserInfo userInfo = getUserInfo();
        if (userInfo == null || userInfo.getRoles() == null) {
            return new ArrayList<>();
        }
        return userInfo.getRoles().stream().map(SysRole::getCode).toList();
    }

    /**
     * 获取权限集合
     */
    public static List<String> getPermissionNames() {
        UserInfo userInfo = getUserInfo();
        if (userInfo == null || userInfo.getPerms() == null) {
            return new ArrayList<>();
        }
        return userInfo.getPerms().stream().toList();
    }

    /**
     * 密码加密
     */
    public static String encode(String salt, String password) {
        return SaSecureUtil.aesEncrypt(salt, password);
    }

    /**
     * 密码解密
     */
    public static String decrypt(String salt, String password) {
        return SaSecureUtil.aesDecrypt(salt, password);
    }

    public static void main(String[] args) {
        System.out.println(encode("hduchat-salt", "123456"));
    }
}
