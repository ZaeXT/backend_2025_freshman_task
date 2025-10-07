package com.chiyu.hduchat.upms.mapper;

import com.chiyu.hduchat.upms.model.entity.SysDept;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Mapper;

/**
 * 部门表(Dept)表数据库访问层
 *
 * @author chiyu
 * @since 2025/10
 */
@Mapper
public interface SysDeptMapper extends BaseMapper<SysDept> {

}
