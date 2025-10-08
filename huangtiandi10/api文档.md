# API 文档（简明版）
概览

基本路径：/auth（公开） 和 /qa（需要 JWT）

认证：JWT，登录后前端在 Authorization: Bearer <token> 中携带

响应格式：JSON

## 1. POST /auth/register — 注册

说明：创建新用户，初始化 vip_level=0、question_count=0。

请求 Body (JSON):

{
  "username": "alice",
  "password": "plainpassword"
}


成功响应 (200):

{ "message": "注册成功" }


失败:

400 参数错误

409 用户名已存在

500 其他错误

注意：后端应该在 CreateUser 检测 Duplicate entry 并返回更有意义的错误（见代码片段）。

## 2. POST /auth/login — 登录

说明：用户名/密码校验通过后返回 JWT（以及用户 ID / vip_level 可选）

请求 Body (JSON):

{
  "username": "alice",
  "password": "plainpassword"
}


成功响应:

{
  "token": "eyJhbGciOi... ",
  "user_id": 14,
  "vip_level": 0,
  "question_count": 0
}


失败：

400 参数错误

401 用户名或密码错误

## 3. POST /qa/ask — 提问 / 调用 AI

说明：用户提问；后端负责：

从 JWT 获得 userID

根据 user.question_count（或 user.VIPLevel）确定可用模型（用 Casbin 或策略选择）

调用 AI（CallAI(model, question)）

保存一条 QA 记录（保存 user_id, question, answer, model_used）

更新用户 question_count（并视需要更新 vip_level）

请求 Body (JSON):

{
  "question": "乒乓球运动对身体好吗"
}


成功响应 (200):

{
  "question": "...",
  "answer": "...",
  "model": "deepseek-v3.1",
  "question_count": 5   // 返回更新后的次数（可选）
}


可能的错误：

400 参数错误

403 无可用模型（Casbin/策略判断失败）

500 AI 调用或数据库错误

要点：

选模型逻辑建议由 Casbin + vip_level 驱动（见下方代码示例）。

保存记录和更新 question_count 建议在一个事务内完成（避免竞态）。

## 4. GET /qa/history — 获取历史问答

说明：返回用户历史问答记录（不在每条记录里返回 question_count，如果需要可单独返回 top-level question_count）

请求头：

Authorization: Bearer <token>


成功响应 (200) 推荐格式：

{
  "question_count": 12,       // 当前用户总提问次数（可选）
  "records": [
    {
      "id": 13,
      "user_id": 15,
      "question": "2+3=?",
      "answer": "5",
      "model_used": "deepseek-v3.1"
    },
    ...
  ]
}


注意：

前端若不需要 question_count，可以把 question_count 去掉；为方便使用建议将 question_count 单独放 top-level 而不是每条记录内。

## 5. DELETE /qa/delete/:id — 删除单条记录

说明：删除指定 qa_records 的记录（仅能删除自己的记录，需在 controller 加权限校验）。

路径参数：

id：QARecord 的 id

响应：

200 {"message":"删除成功"}

404 {"error":"记录不存在"}

400 {"error":"无效的ID"}

注意：

GORM Delete 不会在找不到记录时返回 error，需要检查 RowsAffected（见代码示例）。

## 6. DELETE /qa/clear — 清空当前用户的历史

说明：删除该用户的所有 QA 记录

响应：最好返回删除的数量或成功消息。例如:

{"deleted": 5}