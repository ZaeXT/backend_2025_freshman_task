package com.chiyu.hduchat.upms.service.impl;

import cn.hutool.core.collection.CollUtil;
import cn.hutool.core.lang.Dict;
import cn.hutool.core.lang.tree.Tree;
import cn.hutool.core.lang.tree.TreeNode;
import cn.hutool.core.lang.tree.TreeUtil;
import com.chiyu.hduchat.common.exception.ServiceException;
import com.chiyu.hduchat.upms.model.entity.SysDept;
import com.chiyu.hduchat.upms.mapper.SysDeptMapper;
import com.chiyu.hduchat.upms.service.SysDeptService;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

/**
 * 部门表(Dept)表服务实现类
 *
 * @author chiyu
 * @since 2025/10
 */
@Service
@RequiredArgsConstructor
public class SysDeptServiceImpl extends ServiceImpl<SysDeptMapper, SysDept> implements SysDeptService {

    @Override
    public List<SysDept> list(SysDept sysDept) {
        return baseMapper.selectList(new LambdaQueryWrapper<SysDept>()
                .orderByAsc(SysDept::getOrderNo));
    }

    @Override
    public List<Tree<Object>> tree(SysDept sysDept) {
        List<SysDept> sysDeptList = baseMapper.selectList(new LambdaQueryWrapper<SysDept>()
                .ne(sysDept.getId() != null, SysDept::getId, sysDept.getId()));

        // 构建树形结构
        List<TreeNode<Object>> nodeList = CollUtil.newArrayList();
        sysDeptList.forEach(t -> {
            TreeNode<Object> node = new TreeNode<>(
                    t.getId(),
                    t.getParentId(),
                    t.getName(),
                    0
            );
            node.setExtra(Dict.create().set("orderNo", t.getOrderNo()).set("des", t.getDes()));
            nodeList.add(node);
        });
        return TreeUtil.build(nodeList, "0");
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void delete(String id) {
        List<SysDept> list = baseMapper.selectList(new LambdaQueryWrapper<SysDept>().eq(SysDept::getParentId, id));
        if (!list.isEmpty()) {
            throw new ServiceException("该部门包含子节点，不能删除");
        }
        baseMapper.deleteById(id);
    }
}
