package com.chiyu.hduchat.upms.service;

import cn.hutool.core.lang.tree.Tree;
import com.chiyu.hduchat.upms.model.entity.SysDept;
import com.baomidou.mybatisplus.extension.service.IService;

import java.util.List;

/**
 * 部门表(Dept)表服务接口
 *
 * @author chiyu
 * @since 2025/10
 */
public interface SysDeptService extends IService<SysDept> {

    List<SysDept> list(SysDept sysDept);

    List<Tree<Object>> tree(SysDept sysDept);

    void delete(String id);

}
