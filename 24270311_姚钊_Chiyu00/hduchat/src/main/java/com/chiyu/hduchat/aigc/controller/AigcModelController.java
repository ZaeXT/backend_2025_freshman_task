package com.chiyu.hduchat.aigc.controller;

import cn.dev33.satoken.annotation.SaCheckPermission;
import cn.hutool.core.util.StrUtil;
import com.chiyu.hduchat.aigc.model.component.ProviderRefreshEvent;
import com.chiyu.hduchat.aigc.model.entity.AigcModel;
import com.chiyu.hduchat.aigc.service.AigcModelService;
import com.chiyu.hduchat.common.annotation.ApiLog;
import com.chiyu.hduchat.configuration.SpringContextHolder;
import com.chiyu.hduchat.common.utils.MybatisUtil;
import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.common.utils.R;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * @author chiyu
 * @since 2025/10
 */
@RestController
@RequiredArgsConstructor
@RequestMapping("/aigc/model")
public class AigcModelController {

    private final AigcModelService modelService;
    private final SpringContextHolder contextHolder;

    @GetMapping("/list")
    public R<List<AigcModel>> list(AigcModel data) {
        return R.ok(modelService.list(data));
    }

    @GetMapping("/page")
    public R list(AigcModel data, QueryPage queryPage) {
        Page<AigcModel> iPage = modelService.page(data, queryPage);
        return R.ok(MybatisUtil.getData(iPage));
    }

    @GetMapping("/{id}")
    public R<AigcModel> findById(@PathVariable String id) {
        return R.ok(modelService.selectById(id));
    }

    @PostMapping
    @ApiLog("添加模型")
    @SaCheckPermission("aigc:model:add")
    public R add(@RequestBody AigcModel data) {
        if (StrUtil.isNotBlank(data.getApiKey()) && data.getApiKey().contains("*")) {
            data.setApiKey(null);
        }
        if (StrUtil.isNotBlank(data.getSecretKey()) && data.getSecretKey().contains("*")) {
            data.setSecretKey(null);
        }
        modelService.save(data);
        SpringContextHolder.publishEvent(new ProviderRefreshEvent(data));
        return R.ok();
    }

    @PutMapping
    @ApiLog("修改模型")
    @SaCheckPermission("aigc:model:update")
    public R update(@RequestBody AigcModel data) {
        if (StrUtil.isNotBlank(data.getApiKey()) && data.getApiKey().contains("*")) {
            data.setApiKey(null);
        }
        if (StrUtil.isNotBlank(data.getSecretKey()) && data.getSecretKey().contains("*")) {
            data.setSecretKey(null);
        }
        modelService.updateById(data);
        SpringContextHolder.publishEvent(new ProviderRefreshEvent(data));
        return R.ok();
    }

    @DeleteMapping("/{id}")
    @ApiLog("删除模型")
    @SaCheckPermission("aigc:model:delete")
    public R delete(@PathVariable String id) {
        modelService.removeById(id);

        // Delete dynamically registered beans, according to ID
        contextHolder.unregisterBean(id);
        return R.ok();
    }
}

