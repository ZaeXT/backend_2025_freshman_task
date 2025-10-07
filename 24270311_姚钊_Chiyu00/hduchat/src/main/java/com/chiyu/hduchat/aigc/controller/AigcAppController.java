package com.chiyu.hduchat.aigc.controller;

import cn.dev33.satoken.annotation.SaCheckPermission;
import cn.hutool.core.lang.Dict;
import com.chiyu.hduchat.aigc.model.entity.AigcApp;
import com.chiyu.hduchat.aigc.service.AigcAppService;
import com.chiyu.hduchat.component.store.AppStore;
import com.chiyu.hduchat.common.annotation.ApiLog;
import com.chiyu.hduchat.common.utils.MybatisUtil;
import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.common.utils.R;
import com.baomidou.mybatisplus.core.toolkit.Wrappers;
import lombok.RequiredArgsConstructor;
import org.apache.commons.lang3.StringUtils;
import org.springframework.web.bind.annotation.*;

import java.util.Date;
import java.util.List;

@RestController
@RequiredArgsConstructor
@RequestMapping("/aigc/app")
public class AigcAppController {

    private final AigcAppService aigcAppService;
    private final AppStore appStore;

    @GetMapping("/list")
    public R<List<AigcApp>> list(AigcApp data) {
        return R.ok(aigcAppService.list(data));
    }

    @GetMapping("/page")
    public R<Dict> page(AigcApp data, QueryPage queryPage) {
        return R.ok(MybatisUtil.getData(aigcAppService.page(MybatisUtil.wrap(data, queryPage),
                Wrappers.<AigcApp>lambdaQuery()
                        .like(StringUtils.isNotEmpty(data.getName()), AigcApp::getName, data.getName())
        )));
    }

    @GetMapping("/{id}")
    public R<AigcApp> findById(@PathVariable String id) {
        AigcApp app = aigcAppService.getById(id);
        return R.ok(app);
    }

    @PostMapping
    @ApiLog("新增应用")
    @SaCheckPermission("aigc:app:add")
    public R add(@RequestBody AigcApp data) {
        data.setCreateTime(new Date());
        data.setSaveTime(new Date());
        aigcAppService.save(data);
        appStore.init();
        return R.ok();
    }

    @PutMapping
    @ApiLog("修改应用")
    @SaCheckPermission("aigc:app:update")
    public R update(@RequestBody AigcApp data) {
        data.setSaveTime(new Date());
        aigcAppService.updateById(data);
        appStore.init();
        return R.ok();
    }

    @DeleteMapping("/{id}")
    @ApiLog("删除应用")
    @SaCheckPermission("aigc:app:delete")
    public R delete(@PathVariable String id) {
        aigcAppService.removeById(id);
        appStore.init();
        return R.ok();
    }
}
