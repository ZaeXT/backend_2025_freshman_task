package com.chiyu.hduchat.upms.utils;

import com.chiyu.hduchat.upms.model.dto.MenuTree;
import com.chiyu.hduchat.upms.model.entity.SysMenu;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.stream.Collectors;

/**
 * @author chiyu
 * @since 2025/10
 */
public class TreeUtil {

    private static List<MenuTree<SysMenu>> init(List<SysMenu> list) {
        List<MenuTree<SysMenu>> treeList = new ArrayList<>();
        list.forEach(menu -> {
            MenuTree<SysMenu> tree = new MenuTree<>();
            tree.setId(menu.getId());
            tree.setParentId(menu.getParentId());
            tree.setName(menu.getName());
            tree.setPath(menu.getPath());
            tree.setType(menu.getType());
            tree.setComponent(menu.getComponent());
            tree.setPerms(menu.getPerms());
            tree.setMeta(new MenuTree.MenuMeta(menu.getName(), menu.getIcon()));
            tree.setOrderNo(menu.getOrderNo());
            tree.setDisabled(menu.getIsDisabled());
            tree.setIsExt(menu.getIsExt());
            tree.setKeepalive(menu.getIsKeepalive());
            tree.setShow(menu.getIsShow());
            treeList.add(tree);
        });
        return treeList.stream().sorted(Comparator.comparing(MenuTree::getOrderNo)).collect(Collectors.toList());
    }


    public static List<MenuTree<SysMenu>> build(List<SysMenu> list) {
        List<MenuTree<SysMenu>> nodes = init(list);
        List<MenuTree<SysMenu>> tree = new ArrayList<>();
        nodes.forEach(node -> {
            String pid = node.getParentId();
            if (node.getIsExt()) {
                node.setComponent("Layout");
                node.setName(node.getPath());
            }
            if (pid == null || pid.equals("0")) {
                // 父级节点
                tree.add(node);
                return;
            }
            for (MenuTree<SysMenu> c : nodes) {
                String id = c.getId();
                if (id != null && id.equals(pid)) {
                    c.getChildren().add(node);
                    return;
                }
            }
        });
        return tree;
    }
}
