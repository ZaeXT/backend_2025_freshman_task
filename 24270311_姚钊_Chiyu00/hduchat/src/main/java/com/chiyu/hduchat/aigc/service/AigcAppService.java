package com.chiyu.hduchat.aigc.service;

import com.chiyu.hduchat.aigc.model.entity.AigcApp;
import com.baomidou.mybatisplus.extension.service.IService;

import java.util.List;

/**
 * @author chiyu
 * @since 2025/10
 */
public interface AigcAppService extends IService<AigcApp> {

    List<AigcApp> list(AigcApp data);

    AigcApp getById(String id);
}
