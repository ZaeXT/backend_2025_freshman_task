package com.chiyu.hduchat.component.store;

import com.chiyu.hduchat.aigc.model.entity.AigcApp;
import com.chiyu.hduchat.aigc.service.AigcAppService;
import jakarta.annotation.PostConstruct;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * @author chiyu
 * @since 2025/10
 */
@Slf4j
@Component
@AllArgsConstructor
public class AppStore {

    private static final Map<String, AigcApp> appMap = new HashMap<>();
    private final AigcAppService aigcAppService;

    @PostConstruct
    public void init() {
        log.info("initialize app config list...");
        List<AigcApp> list = aigcAppService.list();
        list.forEach(i -> appMap.put(i.getId(), i));
    }

    public AigcApp get(String appId) {
        return appMap.get(appId);
    }
}
