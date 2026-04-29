# 生长追踪小程序 - API接口文档

## 基础信息

- **Base URL**: `http://your-domain.com/api/v1`
- **Content-Type**: `application/json`
- **认证方式**: 所有 `/api/v1` 下的接口 (除 `/health` 和 `/auth/login` 和 `/pay/callback` 外) 都需要在请求头中携带:
  ```
  Authorization: Bearer <token>
  ```

---

## 统一响应格式

**所有接口统一返回格式：**

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | **必须判断此值**：0=成功，非0=失败 |
| msg | string | 成功或错误描述 |
| data | object/array/null | 响应数据 |

> **前端必须判断：`code === 0 才是成功！**

### 状态码说明

| code | 说明 |
|------|------|
| 0 | 成功 |
| 1 | 服务器内部错误 |
| 2 | 参数错误 |
| 3 | 未授权/未登录 |
| 4 | 资源不存在 |
| 5 | 配额不足 |
| 6 | 邀请码无效 |
| 7 | 家庭已满 |

### HTTP状态码说明

| HTTP状态码 | 说明 |
|------------|------|
| 200 | 成功 (GET/PUT/DELETE 成功) |
| 201 | 创建成功 (POST 创建成功) |
| 400 | 参数错误 |
| 401 | 未授权 |
| 404 | 不存在 |
| 500 | 服务器错误 |

> ❌ **错误的判断方式：
```javascript
if (res.statusCode === 200) // 不要这样写！201也是成功
```

✅ **正确的判断方式：**
```javascript
if (res.data.code === 0) {
  // 成功
} else {
  // 失败，显示 res.data.msg
}
```

---

## 接口列表

---

### 1. 健康检查

```
GET /health
```

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "status": "healthy",
    "version": "1.0.0"
  }
}
```

---

### 2. 认证接口

#### 2.1 微信登录

```
POST /auth/login
```

**请求体**:
```json
{
  "code": "微信登录code"
}
```

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "token": "jwt_token_用户ID_时间戳",
    "expire_at": 1714400000,
    "user": {
      "id": "xxx",
      "nickname": "昵称"
    }
  }
}
```

---

### 3. 用户接口 (需要认证)

#### 3.1 获取用户信息

```
GET /user/info
```

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": "xxx",
    "nickname": "昵称",
    "avatar_url": "头像URL",
    "phone": "手机号"
  }
}
```

#### 3.2 更新用户信息

```
PUT /user/info
```

**请求体**:
```json
{
  "nick_name": "昵称",
  "avatar_url": "头像URL",
  "phone": "手机号"
}
```

**响应**:
```json
{
  "code": 0,
  "msg": "更新成功"
}
```

---

### 4. 宝宝管理 (需要认证)

#### 4.1 获取宝宝列表

```
GET /children
```

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "id": "xxx",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "user_id": "xxx",
      "family_id": 1,
      "nickname": "宝宝昵称",
      "gender": "male",
      "birthday": "2020-01-01T00:00:00Z",
      "initial_height": 50.0,
      "initial_weight": 3.5,
      "father_height": 175.0,
      "mother_height": 165.0,
      "standard_type": "cn"
    }
  ]
}
```

#### 4.2 创建宝宝

```
POST /children
```

**请求体**:
```json
{
  "name": "宝宝昵称",
  "gender": "male",
  "birthday": "2020-01-01",
  "father_height": 175,
  "mother_height": 165
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| name | 是 | 宝宝昵称，最多64字符 |
| gender | 是 | male 或 female |
| birthday | 是 | 格式 YYYY-MM-DD |
| father_height | 是 | 100-250 cm |
| mother_height | 是 | 100-250 cm |

**响应** (HTTP 201):
```json
{
  "code": 0,
  "msg": "创建成功",
  "data": {
    "id": "xxx",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "user_id": "xxx",
    "nickname": "宝宝昵称",
    "gender": "male",
    "birthday": "2020-01-01T00:00:00Z",
    "father_height": 175.0,
    "mother_height": 165.0
  }
}
```

> 注意：这里返回 HTTP 201，不是 200！但 `code === 0` 就是成功。

#### 4.3 获取宝宝详情

```
GET /children/:id
```

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": "xxx",
    "nickname": "宝宝昵称",
    ...
  }
}
```

#### 4.4 更新宝宝

```
PUT /children/:id
```

**请求体**: 同创建

**响应**:
```json
{
  "code": 0,
  "msg": "更新成功"
}
```

#### 4.5 删除宝宝

```
DELETE /children/:id
```

**响应**:
```json
{
  "code": 0,
  "msg": "删除成功"
}
```

#### 4.6 切换当前宝宝

```
POST /children/switch
```

**请求体**:
```json
{
  "child_id": "xxx"
}
```

**响应**:
```json
{
  "code": 0,
  "msg": "切换成功"
}
```

---

### 5. 生长记录 (需要认证)

#### 5.1 获取记录列表

```
GET /records?child_id=xxx&page=1&page_size=20
```

| 参数 | 必填 | 说明 |
|------|------|------|
| child_id | **是** | 宝宝ID，**必须传** |
| start_date | 否 | 开始日期 YYYY-MM-DD |
| end_date | 否 | 结束日期 YYYY-MM-DD |
| page | 否 | 页码，默认1 |
| page_size | 否 | 每页数量，默认20 |

> ❌ **不传 child_id 会返回 400 + code=2

**成功响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "items": [
      {
        "id": "xxx",
        "child_id": "xxx",
        "measure_date": "2024-01-01T00:00:00Z",
        "height": 120.5,
        "weight": 25.5,
        "note": "备注",
        "photo": "图片URL"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.2 创建记录

```
POST /records
```

**请求体**:
```json
{
  "child_id": "xxx",
  "height": 120.5,
  "weight": 25.5,
  "date": "2024-01-01",
  "note": "备注",
  "photo": "图片URL"
}
```

**响应** (HTTP 201):
```json
{
  "code": 0,
  "msg": "创建成功",
  "data": {
    "id": "xxx",
    ...
  }
}
```

#### 5.3 更新记录

```
PUT /records/:id
```

**请求体**: 同创建 (不含 child_id

**响应**:
```json
{
  "code": 0,
  "msg": "更新成功"
}
```

#### 5.4 删除记录

```
DELETE /records/:id
```

**响应**:
```json
{
  "code": 0,
  "msg": "删除成功"
}
```

---

### 6. 家庭管理 (需要认证)

#### 6.1 获取家庭信息

```
GET /family
```

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "name": "家庭名称",
    "invite_code": "ABC123",
    "member_count": 3,
    "child_count": 2
  }
}
```

#### 6.2 创建家庭

```
POST /family
```

**请求体**:
```json
{
  "name": "家庭名称"
}
```

#### 6.3 加入家庭

```
POST /family/join
```

**请求体**:
```json
{
  "invite_code": "ABC123",
  "role": "editor"
}
```

| role 可选值: editor, viewer

#### 6.4 退出家庭

```
DELETE /family/leave
```

#### 6.5 更新成员角色

```
PUT /family/members/:id/role
```

**请求体**:
```json
{
  "member_id": "xxx",
  "role": "editor"
}
```

#### 6.6 生成邀请码

```
POST /family/inviteCode
```

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "invite_code": "NEWCODE",
    "share_url": "分享链接"
  }
}
```

---

### 7. 订阅/会员 (需要认证)

#### 7.1 获取订阅状态

```
GET /subscription
```

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "is_active": true,
    "remaining_quota": 100,
    "member_benefits": [
      {"icon": "xxx", "text": "权益1"}
    ]
  }
}
```

#### 7.2 创建支付订单

```
POST /subscription/createOrder
```

**请求体**:
```json
{
  "code": "微信登录code",
  "plan_id": "monthly",
  "product_id": "xxx",
  "total_fee": 9900
}
```

| plan_id 可选值: monthly, quarterly, yearly
| total_fee 单位：分

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "timeStamp": "xxx",
    "nonceStr": "xxx",
    "package": "prepay_id=xxx",
    "signType": "RSA",
    "paySign": "xxx",
    "order_id": "xxx"
  }
}
```

---

### 8. AI功能 (需要认证)

#### 8.1 AI对话

```
POST /ai/chat
```

**请求体**:
```json
{
  "child_id": "xxx",
  "message": "我家宝宝身高正常吗？",
  "context": [
    {"role": "user", "content": "你好"}
  ]
}
```

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "response": "AI回复内容",
    "tokens": 100
  }
}
```

#### 8.2 解析体检报告

```
POST /ai/parseReport
```

**请求体**:
```json
{
  "child_id": "xxx",
  "image_url": "图片URL",
  "report_type": "growth"
}
```

**响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "ocr_text": "OCR识别的文本",
    "ai_result": {
      "key_indicators": [],
      "normal_ranges": {}
    }
  }
}
```

---

### 9. 首页数据 (需要认证)

```
GET /home
```

---

### 10. 支付回调 (微信服务器调用，前端不需要管)

```
POST /pay/callback
```

---

## 日期格式说明

| 场景 | 格式 | 例子 |
|------|------|------|
| 你发送给后端 (birthday, date) | `YYYY-MM-DD` | `"2020-01-01" |
| 后端返回给你 (created_at, updated_at, birthday) | ISO 8601 | `"2020-01-01T00:00:00Z" |

---

## 前端统一错误处理

所有接口失败时都会返回：

```json
{
  "code": 2,
  "msg": "参数错误: 缺少child_id参数",
  "data": null
}
```

**处理方式：
```javascript
if (res.data.code !== 0) {
  wx.showToast({
    title: res.data.msg,
    icon: 'none'
  })
  return
}
```

---

## Token格式

**登录成功后保存：
```javascript
wx.setStorageSync('token', 登录返回的token)
```

**每次请求带上：**
```javascript
header: {
  'Authorization': 'Bearer ' + wx.getStorageSync('token')
}
```

---

## 已知问题 & 解决

| 问题 | 原因 | 解决 |
|------|------|------|
| 创建宝宝提示"创建成功 Error" | HTTP返回201，前端只判断200 | 判断 `code === 0` |
| /records 返回400 | 没传 child_id | 必须传 `/records?child_id=xxx` |

---

**最后更新**: 2024-04-29
