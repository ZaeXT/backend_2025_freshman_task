package com.chiyu.hduchat.upms.service;

import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.upms.model.dto.UserInfo;
import com.chiyu.hduchat.upms.model.entity.SysUser;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.service.IService;

import java.util.List;

/**
 * 用户表(User)表服务接口
 *
 * @author chiyu
 * @since 2025/10
 */
public interface SysUserService extends IService<SysUser> {

    /**
     * 根据用户名查询
     */
    SysUser findByName(String username);

    /**
     * 根据ID查询
     */
    UserInfo findById(String userId);

    /**
     * 查询用户数据
     */
    UserInfo info(String username);

    /**
     * 条件查询
     */
    List<SysUser> list(SysUser sysUser);

    /**
     * 分页、条件查询
     */
    IPage<UserInfo> page(UserInfo user, QueryPage queryPage);

    /**
     * 校验用户名是否存在
     */
    boolean checkName(UserInfo sysUser);

    void add(UserInfo user);

    void update(UserInfo user);

    void delete(String id, String username);

    void reset(String id, String password, String username);
}
