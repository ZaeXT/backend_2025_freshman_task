package com.chiyu.hduchat.aigc.service;

import com.chiyu.hduchat.aigc.model.entity.AigcModel;
import com.chiyu.hduchat.common.utils.QueryPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.baomidou.mybatisplus.extension.service.IService;

import java.util.List;

/**
 * @author chiyu
 * @since 2025/10
 */
public interface AigcModelService extends IService<AigcModel> {

    List<AigcModel> getChatModels();

    List<AigcModel> getImageModels();

    List<AigcModel> getEmbeddingModels();

    List<AigcModel> list(AigcModel data);

    Page<AigcModel> page(AigcModel data, QueryPage queryPage);

    AigcModel selectById(String id);
}

