package com.chiyu.hduchat.aigc.service.impl;

import cn.hutool.core.util.StrUtil;
import com.chiyu.hduchat.aigc.model.entity.AigcModel;
import com.chiyu.hduchat.aigc.service.AigcAppService;
import com.chiyu.hduchat.aigc.service.AigcModelService;
import com.chiyu.hduchat.aigc.model.entity.AigcApp;
import com.chiyu.hduchat.aigc.mapper.AigcAppMapper;
import com.baomidou.mybatisplus.core.toolkit.Wrappers;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

/**
 * @author chiyu
 * @since 2025/10
 */
@RequiredArgsConstructor
@Service
public class AigcAppServiceImpl extends ServiceImpl<AigcAppMapper, AigcApp> implements AigcAppService {

    private final AigcModelService aigcModelService;

    @Override
    public List<AigcApp> list(AigcApp data) {
        List<AigcApp> list = baseMapper.selectList(Wrappers.<AigcApp>lambdaQuery()
                .like(StrUtil.isNotBlank(data.getName()), AigcApp::getName, data.getName()));

        Map<String, List<AigcModel>> modelMap = aigcModelService.list(new AigcModel()).stream().collect(Collectors.groupingBy(AigcModel::getId));
        list.forEach(i -> {
            List<AigcModel> models = modelMap.get(i.getModelId());
            if (models != null) {
                i.setModel(models.get(0));
            }
        });
        return list;
    }

    @Override
    public AigcApp getById(String id) {
        AigcApp app = baseMapper.selectById(id);
        if (app != null) {
            String modelId = app.getModelId();
            if (modelId != null) {
                app.setModel(aigcModelService.selectById(modelId));
            }
        }
        return app;
    }
}
