package com.chiyu.hduchat.component.log;

import org.springframework.context.ApplicationEvent;

/**
 * 自定义定义 Log事件
 *
 * @author chiyu
 * @since 2025/10
 */
public class LogEvent extends ApplicationEvent {

    public LogEvent(Object source) {
        super(source);
    }
}
