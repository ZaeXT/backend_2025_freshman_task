package com.chiyu.hduchat.aigc.controller;

import cn.hutool.core.lang.Dict;
import com.chiyu.hduchat.aigc.mapper.AigcAppMapper;
import com.chiyu.hduchat.aigc.mapper.AigcMessageMapper;
import com.chiyu.hduchat.common.utils.R;
import com.chiyu.hduchat.upms.mapper.SysUserMapper;
import com.baomidou.mybatisplus.core.toolkit.Wrappers;
import lombok.AllArgsConstructor;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * @author chiyu
 * @since 2025/10
 */
@RequestMapping("/aigc/statistic")
@RestController
@AllArgsConstructor
public class AigcStatisticsController {

    private final AigcMessageMapper aigcMessageMapper;
    private final SysUserMapper userMapper;
    private final AigcAppMapper aigcAppMapper;

    @GetMapping("/requestBy30")
    public R request30Chart() {
        return R.ok(aigcMessageMapper.getReqChartBy30());
    }

    @GetMapping("/tokenBy30")
    public R token30Chart() {
        return R.ok(aigcMessageMapper.getTokenChartBy30());
    }

    @GetMapping("/token")
    public R tokenChart() {
        return R.ok(aigcMessageMapper.getTokenChart());
    }

    @GetMapping("/request")
    public R requestChart() {
        return R.ok(aigcMessageMapper.getReqChart());
    }

    @GetMapping("/home")
    public R home() {
        Dict reqData = aigcMessageMapper.getCount();
        Dict totalData = aigcMessageMapper.getTotalSum();
        Dict userData = userMapper.getCount();
        Long totalPrompt = aigcAppMapper.selectCount(Wrappers.query());
        Dict result = Dict.create();
        result.putAll(reqData);
        result.putAll(totalData);
        result.putAll(userData);
        result.set("totalPrompt", totalPrompt.intValue());
        return R.ok(result);
    }
}
