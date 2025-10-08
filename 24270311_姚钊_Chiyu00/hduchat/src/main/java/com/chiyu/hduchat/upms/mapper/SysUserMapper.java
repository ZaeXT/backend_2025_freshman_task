package com.chiyu.hduchat.upms.mapper;

import cn.hutool.core.lang.Dict;
import com.chiyu.hduchat.upms.model.dto.UserInfo;
import com.chiyu.hduchat.upms.model.entity.SysUser;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;

/**
 * 用户表(User)表数据库访问层
 *
 * @author chiyu
 * @since 2025/10
 */
@Mapper
public interface SysUserMapper extends BaseMapper<SysUser> {

    @Select("""
        SELECT
            COALESCE(COUNT(*), 0) AS totalUser,
            COALESCE(SUM( CASE WHEN YEAR ( create_time ) = YEAR ( CURDATE()) AND MONTH ( create_time ) = MONTH ( CURDATE()) THEN 1 ELSE 0 END ), 0) AS curUser
        FROM
            sys_user;
    """)
    Dict getCount();

    IPage<UserInfo> page(IPage<SysUser> page, UserInfo user, String ignoreId, String ignoreName);
}
