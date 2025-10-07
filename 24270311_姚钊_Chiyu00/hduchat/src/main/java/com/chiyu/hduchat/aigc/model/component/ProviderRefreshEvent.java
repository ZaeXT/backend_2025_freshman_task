package com.chiyu.hduchat.aigc.model.component;

import org.springframework.context.ApplicationEvent;

/**
 * @author chiyu
 * @since 2025/10
 */
public class ProviderRefreshEvent extends ApplicationEvent {
    private static final long serialVersionUID = 4109980679877560773L;

    public ProviderRefreshEvent(Object source) {
        super(source);
    }
}
