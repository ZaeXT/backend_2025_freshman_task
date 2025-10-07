package com.chiyu.hduchat.component.aicore.utils;

import cn.hutool.core.bean.BeanUtil;
import com.chiyu.hduchat.component.aicore.model.consts.PromptConst;
import dev.langchain4j.model.input.Prompt;
import dev.langchain4j.model.input.PromptTemplate;

import java.util.Map;

/**
 * @author chiyu
 * @since 2025/10
 */
public class PromptUtil {

    public static Prompt build(String message) {
        return new Prompt(message);
    }

    public static Prompt build(String message, String promptText) {
        return new PromptTemplate(promptText + PromptConst.EMPTY).apply(Map.of(PromptConst.QUESTION, message));
    }

    public static Prompt build(String message, String promptText, Object param) {
        Map<String, Object> params = BeanUtil.beanToMap(param, false, true);
        params.put(PromptConst.QUESTION, message);
        return new PromptTemplate(promptText).apply(params);
    }

    public static Prompt buildDocs(String message) {
        return new PromptTemplate(PromptConst.DOCUMENT).apply(Map.of(PromptConst.QUESTION, message));
    }
}
