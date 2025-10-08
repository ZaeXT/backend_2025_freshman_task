package com.chiyu.hduchat.upms.controller;

import cn.dev33.satoken.annotation.SaCheckPermission;
import cn.dev33.satoken.stp.SaTokenInfo;
import cn.dev33.satoken.stp.StpUtil;
import cn.hutool.core.date.DatePattern;
import cn.hutool.core.date.DateUtil;
import cn.hutool.core.lang.Dict;
import cn.hutool.core.util.StrUtil;
import com.chiyu.hduchat.component.log.SysLogUtil;
import com.chiyu.hduchat.component.auth.model.TokenInfo;
import com.chiyu.hduchat.common.constant.CacheConst;
import com.chiyu.hduchat.common.exception.ServiceException;
import com.chiyu.hduchat.common.properties.AuthProps;
import com.chiyu.hduchat.common.utils.MybatisUtil;
import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.common.utils.R;
import com.chiyu.hduchat.upms.model.dto.UserInfo;
import com.chiyu.hduchat.upms.model.entity.SysRole;
import com.chiyu.hduchat.upms.model.entity.SysUser;
import com.chiyu.hduchat.upms.service.SysRoleService;
import com.chiyu.hduchat.upms.service.SysUserService;
import com.chiyu.hduchat.component.auth.AuthUtil;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.core.toolkit.Wrappers;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import io.swagger.v3.oas.annotations.Operation;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.BeanUtils;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.*;

import static com.chiyu.hduchat.common.constant.CacheConst.AUTH_SESSION_PREFIX;

/**
 * @author chiyu
 * @since 2025/10
 */
@Slf4j
@RestController
@AllArgsConstructor
@RequestMapping("/auth")
public class AuthController {

    private final SysUserService userService;
    private final SysRoleService roleService;
    private final AuthProps authProps;
    private final StringRedisTemplate redisTemplate;

    @PostMapping("/login")
    @Operation(summary = "用户登录", description = "登录成功后获取 Token，用于后续接口认证")
    public R login(@RequestBody UserInfo user ) {
        //https://www.doubao.com/chat/collection/23344863449227522?type=Thread
        if (StrUtil.isBlank(user.getUsername()) || StrUtil.isBlank(user.getPassword())) {
            throw new ServiceException("用户名或密码为空");
        }

        UserInfo userInfo = userService.info(user.getUsername());
        if (userInfo == null) {
            throw new ServiceException("用户名或密码错误");
        }
        if (!AuthUtil.ADMINISTRATOR.equals(userInfo.getUsername()) && !userInfo.getStatus()) {
            throw new ServiceException("该用户已经禁用，请联系管理员");
        }

        String decryptPass = AuthUtil.decrypt(authProps.getSaltKey(), userInfo.getPassword());
        if (!decryptPass.equals(user.getPassword())) {
            throw new ServiceException("密码不正确");
        }

        //todo 1. 参数校验（确保用户唯一标识非空）
        StpUtil.login(userInfo.getId());

        //todo 4.1 维护会话信息（Session 数据存储）
        SaTokenInfo tokenInfo = StpUtil.getTokenInfo();
        StpUtil.getSession()
                .set(CacheConst.AUTH_USER_INFO_KEY, userInfo)
                .set(CacheConst.AUTH_TOKEN_INFO_KEY, tokenInfo);

        SysLogUtil.publish(1, "服务端登录", AuthUtil.getUsername());
        log.info("====> login success，token={}", tokenInfo.getTokenValue());
        //todo 6. 向客户端返回 Token（包含过期时间）
        return R.ok(new TokenInfo().setToken(tokenInfo.tokenValue).setExpiration(tokenInfo.tokenTimeout));
    }

    @DeleteMapping("/logout")
    public R logout() {
        StpUtil.logout();
        return R.ok();
    }

    @PostMapping("/register")
    public R emailRegister(@RequestBody SysUser data) {
        if (StrUtil.isBlank(data.getUsername()) || StrUtil.isBlank(data.getPassword())) {
            throw new ServiceException("用户名或密码为空");
        }

        // 校验用户名是否已存在
        List<SysUser> list = userService.list(Wrappers.<SysUser>lambdaQuery().eq(SysUser::getUsername, data.getUsername()));
        if (!list.isEmpty()) {
            throw new ServiceException("该用户名已存在");
        }

        List<SysRole> roles = roleService.list(Wrappers.<SysRole>lambdaQuery().eq(SysRole::getCode, AuthUtil.DEFAULT_ROLE));
        if (roles.isEmpty()) {
            throw new ServiceException("系统角色配置异常，请联系管理员");
        }

        UserInfo user = (UserInfo) new UserInfo()
                .setRoleIds(roles.stream().map(SysRole::getId).toList())
                .setUsername(data.getUsername())
                .setPassword(AuthUtil.encode(authProps.getSaltKey(), data.getPassword()))
                .setRealName(data.getUsername())
                .setPhone(data.getPhone())
                .setStatus(true)
                .setCreateTime(new Date());
        userService.add(user);
        SysLogUtil.publish(1, "服务端注册", user.getUsername());
        return R.ok();
    }

    @GetMapping("/info")
    public R<UserInfo> info() {
        UserInfo userInfo = userService.info(AuthUtil.getUsername());
        userInfo.setPassword(null);
        return R.ok(userInfo);
    }

    @DeleteMapping("/token/{token}")
    @SaCheckPermission("auth:delete")
    public R tokenDel(@PathVariable String token) {
        StpUtil.kickoutByTokenValue(token);
        return R.ok();
    }

    @GetMapping("/token/page")
    public R tokenPage(QueryPage queryPage) {
        List<String> list = StpUtil.searchTokenValue("", queryPage.getPage() - 1, queryPage.getLimit(), true);
        List ids = redisTemplate.opsForValue().multiGet(list);
        Set<String> keys = redisTemplate.keys(AUTH_SESSION_PREFIX + "*");

        List<Object> result = new ArrayList<>();
        ids.forEach(id -> {
            Dict data = Dict.create();
            Map<String, Object> dataMap = StpUtil.getSessionByLoginId(id).getDataMap();
            UserInfo userInfo = new UserInfo();
            Object obj = dataMap.get(CacheConst.AUTH_USER_INFO_KEY);
            BeanUtils.copyProperties(obj, userInfo);
            if (Objects.equals(AuthUtil.getUserId(), userInfo.getId())) {
                return;
            }
            SaTokenInfo tokenInfo = (SaTokenInfo)dataMap.get(CacheConst.AUTH_TOKEN_INFO_KEY);
            if (tokenInfo == null) {
                return;
            }
            data.set("token", tokenInfo.tokenValue);
            data.set("perms", userInfo.getPerms());
            data.set("roles", userInfo.getRoles());
            data.set("email", userInfo.getEmail());
            data.set("id", userInfo.getId());
            data.set("username", userInfo.getUsername());
            data.set("realName", userInfo.getRealName());

            long expiration = StpUtil.getTokenTimeout();
            Date targetDate = new Date(System.currentTimeMillis() + expiration);
            String formattedDate = DateUtil.format(targetDate, DatePattern.NORM_DATETIME_PATTERN);
            data.set("expiration", formattedDate);

            result.add(data);
        });

        IPage page = new Page(queryPage.getPage(), queryPage.getLimit());
        page.setRecords(result);
        page.setTotal(keys == null ? 0 : keys.size());
        return R.ok(MybatisUtil.getData(page));
    }
}
