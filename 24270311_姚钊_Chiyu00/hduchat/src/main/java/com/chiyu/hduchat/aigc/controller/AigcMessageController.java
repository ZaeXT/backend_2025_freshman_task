package com.chiyu.hduchat.aigc.controller;

import cn.dev33.satoken.annotation.SaCheckPermission;
import cn.hutool.core.util.StrUtil;
import com.chiyu.hduchat.aigc.model.entity.AigcMessage;
import com.chiyu.hduchat.aigc.service.AigcMessageService;
import com.chiyu.hduchat.common.annotation.ApiLog;
import com.chiyu.hduchat.common.utils.MybatisUtil;
import com.chiyu.hduchat.common.utils.QueryPage;
import com.chiyu.hduchat.common.utils.R;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.core.toolkit.Wrappers;
import lombok.AllArgsConstructor;
import org.springframework.web.bind.annotation.*;

/**
 * @author chiyu
 * @since 2025/10
 */
@RequestMapping("/aigc/message")
@RestController
@AllArgsConstructor
public class AigcMessageController {

    private final AigcMessageService aigcMessageService;

    @GetMapping("/page")
    public R list(AigcMessage data, QueryPage queryPage) {
        LambdaQueryWrapper<AigcMessage> queryWrapper = Wrappers.<AigcMessage>lambdaQuery()
                .like(!StrUtil.isBlank(data.getMessage()), AigcMessage::getMessage, data.getMessage())
                .like(!StrUtil.isBlank(data.getUsername()), AigcMessage::getUsername, data.getUsername())
                .eq(!StrUtil.isBlank(data.getRole()), AigcMessage::getRole, data.getRole())
                .orderByDesc(AigcMessage::getCreateTime);
        IPage<AigcMessage> iPage = aigcMessageService.page(MybatisUtil.wrap(data, queryPage), queryWrapper);
        return R.ok(MybatisUtil.getData(iPage));
    }

    @GetMapping("/{id}")
    public R getById(@PathVariable String id) {
        return R.ok(aigcMessageService.getById(id));
    }

    @DeleteMapping("/{id}")
    @ApiLog("删除会话消息")
    @SaCheckPermission("aigc:message:delete")
    public R del(@PathVariable String id) {
        return R.ok(aigcMessageService.removeById(id));
    }

}
