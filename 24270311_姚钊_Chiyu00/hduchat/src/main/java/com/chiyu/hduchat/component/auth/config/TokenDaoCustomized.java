package com.chiyu.hduchat.component.auth.config;

import cn.dev33.satoken.session.SaSession;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
//todo 4.2 自定义 Session 序列化（忽略无关属性）
/**
 * Jackson 定制版 SaSession，忽略 timeout 等属性的序列化
 *
 * @author chiyu
 * @since 1.34.0
 */
@JsonIgnoreProperties({"timeout"})//// 序列化时忽略timeout属性
public class TokenDaoCustomized extends SaSession {
    private static final long serialVersionUID = -7600983549653130681L;

    public TokenDaoCustomized() {
        super();
    }

    /**
     * 构建一个Session对象
     *
     * @param id Session的id
     */
    public TokenDaoCustomized(String id) {
        super(id);
    }

}
