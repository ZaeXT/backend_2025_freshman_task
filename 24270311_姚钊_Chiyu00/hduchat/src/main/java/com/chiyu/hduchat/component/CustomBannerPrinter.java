package com.chiyu.hduchat.component;

import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.ApplicationListener;
import org.springframework.stereotype.Component;

/**
 * 启动横幅组件
 * @author chiyu
 * @since 2025/10
 */
@Slf4j
@Component
public class CustomBannerPrinter implements ApplicationListener<ApplicationReadyEvent> {

    @Override
    public void onApplicationEvent(ApplicationReadyEvent event) {
        System.out.println("""
                """);

        log.info("项目启动完成...... 当前环境：{}", event.getApplicationContext().getEnvironment().getActiveProfiles());
    }
}
