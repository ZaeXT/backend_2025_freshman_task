package com.chiyu.hduchat.upms.mapper;

import com.chiyu.hduchat.upms.model.entity.SysLog;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Mapper;

/**
 * 系统日志表(Log)表数据库访问层
 *
 * @author chiyu
 * @since 2025/10
 */
@Mapper
public interface SysLogMapper extends BaseMapper<SysLog> {

}
