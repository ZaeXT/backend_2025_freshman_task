package com.chiyu.hduchat.upms.service;

import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.upms.model.entity.SysLog;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.service.IService;

/**
 * 系统日志表(Log)表服务接口
 *
 * @author chiyu
 * @since 2025/10
 */
public interface SysLogService extends IService<SysLog> {

    /**
     * 分页、条件查询
     */
    IPage<SysLog> list(SysLog sysLog, QueryPage queryPage);

    /**
     * 新增
     */
    void add(SysLog sysLog);
    /**
     * 删除
     */
    void delete(String id);
}
