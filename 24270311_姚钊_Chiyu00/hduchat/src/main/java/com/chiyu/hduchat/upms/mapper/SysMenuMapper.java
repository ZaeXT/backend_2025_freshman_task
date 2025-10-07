package com.chiyu.hduchat.upms.mapper;

import com.chiyu.hduchat.upms.model.entity.SysMenu;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

/**
 * 菜单表(Menu)表数据库访问层
 *
 * @author chiyu
 * @since 2025/10
 */
@Mapper
public interface SysMenuMapper extends BaseMapper<SysMenu> {

    List<SysMenu> build(@Param("roleIds") List<String> roleIds, @Param("type") String type);
}
