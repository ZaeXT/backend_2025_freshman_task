package com.chiyu.hduchat.configuration;

import io.swagger.v3.oas.models.Components;
import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Contact;
import io.swagger.v3.oas.models.info.Info;
import io.swagger.v3.oas.models.security.SecurityRequirement;
import io.swagger.v3.oas.models.security.SecurityScheme;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.ResourceHandlerRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

/**
 * 跨域配置组件
 * @author chiyu
 * @since 2025/10
 */
@Configuration
public class GlobalMVCConfigurer implements WebMvcConfigurer {



    /**
     * 创建了一个api接口的分组
     * 除了配置文件方式创建分组，也可以通过注册bean创建分组
     */
//    @Bean
//    public GroupedOpenApi adminApi() {
//        return GroupedOpenApi.builder()
//                // 分组名称
//                .group("app-api")
//                // 接口请求路径规则
//                .pathsToMatch("/**")
//                .build();
//    }

    /**
     * 配置基本信息
     */
    @Bean
    public OpenAPI openAPI() {
        return new OpenAPI()
                .info(new Info()
                        // 标题
                        .title("hduchat Api接口文档")
                        // 描述Api接口文档的基本信息
                        .description("hduchat后端接口服务")
                        // 版本
                        .version("v1.0.0")
                        // 设置OpenAPI文档的联系信息，姓名，邮箱。
                        .contact(new Contact().name("chiyu").email("2838488675@qq.com"))
                        // 设置OpenAPI文档的许可证信息，包括许可证名称为"Apache 2.0"，许可证URL为"http://springdoc.org"。
//                        .license(new License().name("Apache 2.0").url("http://springdoc.org"))
                );
    }

    @Bean
    public OpenAPI customOpenAPI() {
        // 定义安全方案名称（与 Token 头对应）
        String securitySchemeName = "Authorization";

        return new OpenAPI()
                // 全局应用该认证方案（所有接口默认需要认证）
                .addSecurityItem(new SecurityRequirement().addList(securitySchemeName))
                .components(new Components()
                        .addSecuritySchemes(securitySchemeName,
                                new SecurityScheme()
                                        .name(securitySchemeName)
                                        .type(SecurityScheme.Type.HTTP)
                                        .scheme("bearer") // 对应 Bearer Token
                                        .bearerFormat("JWT") // Token 格式
                                        .description("登录后获取的 Token，格式：Bearer {token}")
                        )
                );
    }

    /**
     * 设置静态资源映射
     */
    @Override
    public void addResourceHandlers(ResourceHandlerRegistry registry) {
        // 添加静态资源映射规则
        registry.addResourceHandler("/static/**").addResourceLocations("classpath:/static/");
        //配置 knife4j 的静态资源请求映射地址
        registry.addResourceHandler("/doc.html")
                .addResourceLocations("classpath:/META-INF/resources/");
        registry.addResourceHandler("/webjars/**")
                .addResourceLocations("classpath:/META-INF/resources/webjars/");
    }
}
