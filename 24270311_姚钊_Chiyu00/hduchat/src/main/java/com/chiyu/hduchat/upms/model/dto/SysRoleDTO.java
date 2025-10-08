package com.chiyu.hduchat.upms.model.dto;

import com.chiyu.hduchat.upms.model.entity.SysRole;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.experimental.Accessors;

import java.util.List;

/**
 * SysRole DTO封装
 *
 * @author chiyu
 * @since 2025/10
 */
@Data
@Accessors(chain = true)
@EqualsAndHashCode(callSuper = true)
public class SysRoleDTO extends SysRole {
    private static final long serialVersionUID = -5792577217091151552L;

    /**
     * 菜单ID集合
     */
    private List<String> menuIds;
}
