# Growth Tracker Pro API Documentation

**API Version**: v1.0
**Base URL**: `https://api.growth-tracker.com/api/v1`
**Authentication**: Bearer Token (JWT)

---

## 目录

1. [认证](#认证)
2. [用户](#用户)
3. [宝宝](#宝宝)
4. [记录](#记录)
5. [订阅](#订阅)
6. [家庭](#家庭)
7. [AI](#ai)
8. [首页](#首页)
9. [医院](#医院)
10. [支付回调](#支付回调)

---

## 认证

### 微信登录

用户通过微信登录获取Token。

**请求**

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "code": "微信登录code"
}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "token": "jwt_token_xxx",
    "expire_at": 1745827200,
    "user": {
      "id": "xxx",
      "open_id": "xxx",
      "nick_name": "xxx",
      "avatar_url": "xxx"
    }
  }
}
```

---

## 用户

### 获取用户信息

**请求**

```http
GET /api/v1/user/info
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": "xxx",
    "open_id": "xxx",
    "nick_name": "用户昵称",
    "avatar_url": "https://xxx.png",
    "phone": "138****8888",
    "family_id": "xxx",
    "subscription": {
      "plan": "yearly",
      "status": "active",
      "end_date": "2027-04-28"
    }
  }
}
```

### 更新用户信息

**请求**

```http
PUT /api/v1/user/info
Authorization: Bearer {token}
Content-Type: application/json

{
  "nick_name": "新昵称",
  "avatar_url": "https://xxx.png",
  "phone": "13812345678"
}
```

**响应**

```json
{
  "code": 0,
  "msg": "更新成功"
}
```

---

## 宝宝

### 获取宝宝列表

**请求**

```http
GET /api/v1/children
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "id": "xxx",
      "name": "宝宝姓名",
      "gender": "male",
      "birthday": "2022-01-15",
      "father_height": 175.0,
      "mother_height": 165.0,
      "is_active": true
    }
  ]
}
```

### 创建宝宝

**请求**

```http
POST /api/v1/children
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "宝宝姓名",
  "gender": "male",
  "birthday": "2022-01-15",
  "father_height": 175.0,
  "mother_height": 165.0
}
```

**响应**

```json
{
  "code": 0,
  "msg": "创建成功",
  "data": {
    "id": "xxx",
    "name": "宝宝姓名",
    "gender": "male",
    "birthday": "2022-01-15",
    "father_height": 175.0,
    "mother_height": 165.0,
    "is_active": true
  }
}
```

### 获取宝宝详情

**请求**

```http
GET /api/v1/children/:id
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": "xxx",
    "name": "宝宝姓名",
    "gender": "male",
    "birthday": "2022-01-15",
    "father_height": 175.0,
    "mother_height": 165.0,
    "age_str": "4岁3月",
    "target_height": {
      "target_height": 172.0,
      "min_height": 164.0,
      "max_height": 180.0
    },
    "percentile": 50,
    "growth_status": "normal",
    "intervention_window": {
      "start": "2032-01-15",
      "end": "2037-01-15",
      "remaining_days": 3650,
      "is_in_window": false
    },
    "latest_record": {
      "id": "xxx",
      "height": 110.5,
      "weight": 18.5,
      "date": "2026-04-01"
    }
  }
}
```

### 更新宝宝信息

**请求**

```http
PUT /api/v1/children/:id
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "新姓名",
  "father_height": 178.0
}
```

**响应**

```json
{
  "code": 0,
  "msg": "更新成功"
}
```

### 删除宝宝

**请求**

```http
DELETE /api/v1/children/:id
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "删除成功"
}
```

### 切换当前宝宝

**请求**

```http
POST /api/v1/children/switch
Authorization: Bearer {token}
Content-Type: application/json

{
  "child_id": "xxx"
}
```

**响应**

```json
{
  "code": 0,
  "msg": "切换成功"
}
```

---

## 记录

### 获取记录列表

**请求**

```http
GET /api/v1/records?child_id=xxx&start_date=2026-01-01&end_date=2026-04-28&page=1&page_size=20
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "items": [
      {
        "id": "xxx",
        "child_id": "xxx",
        "height": 110.5,
        "weight": 18.5,
        "date": "2026-04-01",
        "age_str": "4.2",
        "age_in_days": 1535,
        "note": "春季体检",
        "created_at": "2026-04-01T10:00:00Z"
      }
    ],
    "total": 24,
    "page": 1,
    "page_size": 20
  }
}
```

### 创建记录

**请求**

```http
POST /api/v1/records
Authorization: Bearer {token}
Content-Type: application/json

{
  "child_id": "xxx",
  "height": 110.5,
  "weight": 18.5,
  "date": "2026-04-01",
  "note": "春季体检"
}
```

**响应**

```json
{
  "code": 0,
  "msg": "创建成功",
  "data": {
    "id": "xxx",
    "child_id": "xxx",
    "height": 110.5,
    "weight": 18.5,
    "date": "2026-04-01",
    "age_str": "4.2",
    "age_in_days": 1535
  }
}
```

### 更新记录

**请求**

```http
PUT /api/v1/records/:id
Authorization: Bearer {token}
Content-Type: application/json

{
  "height": 111.0,
  "note": "更新备注"
}
```

**响应**

```json
{
  "code": 0,
  "msg": "更新成功"
}
```

### 删除记录

**请求**

```http
DELETE /api/v1/records/:id
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "删除成功"
}
```

---

## 订阅

### 获取订阅信息

**请求**

```http
GET /api/v1/subscription
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "plan": "yearly",
    "start_date": "2026-01-01",
    "end_date": "2027-01-01",
    "ai_quota": 400,
    "ai_used": 15,
    "remaining_quota": 385,
    "is_active": true,
    "member_benefits": [
      {"icon": "infinity", "text": "无限次AI分析"},
      {"icon": "history", "text": "历史记录永久保存"},
      {"icon": "priority", "text": "优先使用新功能"}
    ]
  }
}
```

### 创建订单

**请求**

```http
POST /api/v1/subscription/createOrder
Authorization: Bearer {token}
Content-Type: application/json

{
  "code": "微信登录code",
  "plan_id": "yearly",
  "product_id": "gtp_yearly",
  "total_fee": 19990
}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "timeStamp": "1745827200",
    "nonceStr": "xxx",
    "package": "prepay_id=xxx",
    "signType": "MD5",
    "paySign": "xxx",
    "order_id": "GT1745827200xxx"
  }
}
```

---

## 家庭

### 获取家庭信息

**请求**

```http
GET /api/v1/family
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "family_id": "xxx",
    "name": "幸福一家",
    "invite_code": "ABC123",
    "max_members": 10,
    "member_count": 3,
    "child_count": 2,
    "members": [
      {"id": "xxx", "user_id": "xxx", "name": "爸爸", "role": "owner"},
      {"id": "xxx", "user_id": "xxx", "name": "妈妈", "role": "editor"}
    ]
  }
}
```

### 创建家庭

**请求**

```http
POST /api/v1/family
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "幸福一家"
}
```

**响应**

```json
{
  "code": 0,
  "msg": "创建成功",
  "data": {
    "family_id": "xxx",
    "name": "幸福一家",
    "invite_code": "ABC123",
    "max_members": 10
  }
}
```

### 加入家庭

**请求**

```http
POST /api/v1/family/join
Authorization: Bearer {token}
Content-Type: application/json

{
  "invite_code": "ABC123",
  "role": "viewer"
}
```

**响应**

```json
{
  "code": 0,
  "msg": "加入成功"
}
```

### 退出家庭

**请求**

```http
DELETE /api/v1/family/leave
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "退出成功"
}
```

### 生成邀请码

**请求**

```http
POST /api/v1/family/inviteCode
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "invite_code": "ABC123",
    "share_url": "pages/family/join?code=ABC123"
  }
}
```

---

## AI

### AI对话

**请求**

```http
POST /api/v1/ai/chat
Authorization: Bearer {token}
Content-Type: application/json

{
  "child_id": "xxx",
  "message": "我家孩子身高发育正常吗？",
  "context": [
    {"role": "user", "content": "之前的问题"},
    {"role": "assistant", "content": "之前的回答"}
  ]
}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "response": "根据您提供的信息...",
    "tokens": 150
  }
}
```

### 解析化验单

**请求**

```http
POST /api/v1/ai/parseReport
Authorization: Bearer {token}
Content-Type: application/json

{
  "child_id": "xxx",
  "image_url": "https://xxx.com/report.jpg",
  "report_type": "bone_age"
}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "ocr_text": "骨龄: 8岁",
    "ai_result": {
      "key_indicators": [
        {"name": "骨龄", "value": "8岁", "status": "normal"}
      ],
      "analysis": "根据化验单分析...",
      "suggestions": ["继续保持均衡的饮食习惯"]
    }
  }
}
```

---

## 首页

### 获取首页数据

**请求**

```http
GET /api/v1/home
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "has_baby": true,
    "baby": {
      "id": "xxx",
      "name": "宝宝姓名",
      "age_str": "4岁3月",
      "target_height": {"target_height": 172.0, "min_height": 164.0, "max_height": 180.0}
    },
    "latest_record": {"height": 110.5, "weight": 18.5, "date": "2026-04-01"},
    "target_height": 172.0,
    "percentile": 50,
    "growth_status": "normal",
    "is_vip": true,
    "ai_remaining": 385,
    "chart_data": {
      "categories": ["01-15", "02-15", "03-15", "04-15"],
      "series": [{"name": "身高", "data": [100.0, 102.5, 105.0, 108.5]}]
    }
  }
}
```

---

## 医院

### 获取医院列表

**请求**

```http
GET /api/v1/hospitals?city=北京&latitude=39.9&longitude=116.4&page=1&page_size=20
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "items": [
      {
        "id": "hospital_001",
        "name": "北京儿童医院",
        "level": "三级甲等",
        "address": "北京市西城区南礼士路1号",
        "latitude": 39.9423,
        "longitude": 116.3562,
        "phone": "010-59616161",
        "pediatric_endo": true,
        "estimated_fee": "500-2000",
        "city": "北京",
        "district": "西城区",
        "distance": 2.5
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

### 获取医院详情

**请求**

```http
GET /api/v1/hospitals/:id
Authorization: Bearer {token}
```

**响应**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": "hospital_001",
    "name": "北京儿童医院",
    "level": "三级甲等",
    "address": "北京市西城区南礼士路1号",
    "latitude": 39.9423,
    "longitude": 116.3562,
    "phone": "010-59616161",
    "logo": "https://xxx.com/logo.jpg",
    "pediatric_endo": true,
    "estimated_fee": "500-2000",
    "city": "北京",
    "district": "西城区",
    "departments": [
      {
        "id": 1,
        "name": "儿童内分泌科",
        "description": "诊治儿童生长发育相关疾病，如矮小症、性早熟等"
      },
      {
        "id": 2,
        "name": "儿童保健科",
        "description": "儿童健康体检和发育评估"
      }
    ]
  }
}
```

---

## 错误码

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 400 | 参数错误 |
| 401 | 未登录 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器错误 |
| 1001 | AI额度用完 |
| 1002 | 非会员 |
| 2001 | 无效邀请码 |
| 2002 | 家庭成员已满 |
