package com.chiyu.hduchat.common.utils;

import lombok.AllArgsConstructor;
import lombok.Data;

import java.io.Serializable;

/**
 * @author chiyu
 * @since 2025/10
 */
@Data
@AllArgsConstructor
public class QueryPage implements Serializable {
    private static final long serialVersionUID = 1L;

    /**
     * 当前页
     */
    private int page = 1;

    /**
     * 每页的记录数
     */
    private int limit = 10;
}
