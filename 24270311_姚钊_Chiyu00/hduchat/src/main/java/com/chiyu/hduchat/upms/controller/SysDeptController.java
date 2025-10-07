package com.chiyu.hduchat.upms.controller;

import cn.dev33.satoken.annotation.SaCheckPermission;
import cn.hutool.core.lang.tree.Tree;
import com.chiyu.hduchat.common.annotation.ApiLog;
import com.chiyu.hduchat.common.utils.R;
import com.chiyu.hduchat.upms.model.entity.SysDept;
import com.chiyu.hduchat.upms.service.SysDeptService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * 部门表(Dept)表控制层
 *
 * @author chiyu
 * @since 2025/10
 */
@RestController
@RequiredArgsConstructor
@RequestMapping("/upms/dept")

public class SysDeptController {

    private final SysDeptService sysDeptService;

    @GetMapping("/list")
    public R<List<SysDept>> list(SysDept sysDept) {
        return R.ok(sysDeptService.list(sysDept));
    }

    @GetMapping("/tree")
    public R<List<Tree<Object>>> tree(SysDept sysDept) {
        return R.ok(sysDeptService.tree(sysDept));
    }

    @GetMapping("/{id}")
    public R<SysDept> findById(@PathVariable String id) {
        return R.ok(sysDeptService.getById(id));
    }

    @PostMapping
    @ApiLog("新增部门")
    @SaCheckPermission("upms:dept:add")
    public R add(@RequestBody SysDept sysDept) {
        sysDept.setParentId(sysDept.getParentId());
        sysDeptService.save(sysDept);
        return R.ok();
    }

    @PutMapping
    @ApiLog("修改部门")
    @SaCheckPermission("upms:dept:update")
    public R update(@RequestBody SysDept sysDept) {
        sysDept.setParentId(sysDept.getParentId());
        sysDeptService.updateById(sysDept);
        return R.ok();
    }

    @DeleteMapping("/{id}")
    @ApiLog("删除部门")
    @SaCheckPermission("upms:dept:delete")
    public R delete(@PathVariable String id) {
        sysDeptService.delete(id);
        return R.ok();
    }
}
