package com.chiyu.hduchat.upms.service.impl;

import com.chiyu.hduchat.common.utils.MybatisUtil;
import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.upms.model.entity.SysLog;
import com.chiyu.hduchat.upms.mapper.SysLogMapper;
import com.chiyu.hduchat.upms.service.SysLogService;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.core.toolkit.Wrappers;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import lombok.RequiredArgsConstructor;
import org.apache.commons.lang3.StringUtils;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/**
 * 系统日志表(Log)表服务实现类
 *
 * @author chiyu
 * @since 2025/10
 */
@Service
@RequiredArgsConstructor
public class SysLogServiceImpl extends ServiceImpl<SysLogMapper, SysLog> implements SysLogService {

    @Override
    public IPage<SysLog> list(SysLog sysLog, QueryPage queryPage) {
        return baseMapper.selectPage(MybatisUtil.wrap(sysLog, queryPage),
                Wrappers.<SysLog>lambdaQuery()
                        .eq(sysLog.getType() != null, SysLog::getType, sysLog.getType())
                        .like(StringUtils.isNotEmpty(sysLog.getOperation()), SysLog::getOperation, sysLog.getOperation())
                        .orderByDesc(SysLog::getCreateTime)
        );
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void add(SysLog sysLog) {
        baseMapper.insert(sysLog);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void delete(String id) {
        baseMapper.deleteById(id);
    }
}
