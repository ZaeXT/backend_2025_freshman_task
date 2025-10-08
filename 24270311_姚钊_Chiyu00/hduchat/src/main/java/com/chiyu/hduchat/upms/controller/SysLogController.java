package com.chiyu.hduchat.upms.controller;

import cn.dev33.satoken.annotation.SaCheckPermission;
import cn.hutool.core.lang.Dict;
import com.chiyu.hduchat.common.utils.MybatisUtil;
import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.common.utils.R;
import com.chiyu.hduchat.upms.model.entity.SysLog;
import com.chiyu.hduchat.upms.service.SysLogService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

/**
 * 系统日志表(Log)表控制层
 *
 * @author chiyu
 * @since 2025/10
 */
@RestController
@RequiredArgsConstructor
@RequestMapping("/upms/log")
public class SysLogController {

    private final SysLogService sysLogService;

    @GetMapping("/page")
    public R<Dict> list(SysLog sysLog, QueryPage queryPage) {
        return R.ok(MybatisUtil.getData(sysLogService.list(sysLog, queryPage)));
    }

    @GetMapping("/{id}")
    public R<SysLog> findById(@PathVariable String id) {
        return R.ok(sysLogService.getById(id));
    }

    @DeleteMapping("/{id}")
    @SaCheckPermission("upms:log:delete")
    public R delete(@PathVariable String id) {
        sysLogService.delete(id);
        return R.ok();
    }
}
