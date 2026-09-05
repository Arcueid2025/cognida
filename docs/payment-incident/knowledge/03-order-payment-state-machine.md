# 订单与支付状态机说明

文档 ID：`PI-KB-003`  
版本：`v1.0`  
生效日期：`2026-01-01`  
适用范围：订单服务与支付服务状态对账

## 1. 合法状态

【PI-KB-003-1】订单状态：`PENDING_PAYMENT`、`PAYING`、`PAID`、`CANCELLED`、`REFUNDING`、`REFUNDED`。支付状态：`INIT`、`PENDING`、`SUCCESS`、`FAILED`、`CLOSED`、`REFUNDED`。

【PI-KB-003-2】正常支付路径为：订单 `PENDING_PAYMENT → PAYING → PAID`，支付 `INIT → PENDING → SUCCESS`。支付成功事件是订单更新为 `PAID` 的唯一自动触发来源。

## 2. 对账规则

【PI-KB-003-3】支付为 `SUCCESS` 且订单为 `PAYING` 或 `PENDING_PAYMENT`，判定为“支付成功但订单未更新”；优先检查回调和消费记录。支付为 `FAILED` 或 `CLOSED` 而订单为 `PAID`，判定为“订单支付状态异常”，禁止自动回退，需人工核对资金流水。

【PI-KB-003-4】订单 `CANCELLED` 后收到支付成功回调，判定为“取消后支付成功”。此时不得直接恢复订单；应先冻结履约，并由人工确认是否恢复订单或发起退款。

【PI-KB-003-5】订单实付金额必须等于支付成功记录金额。金额不一致时，系统不得执行自动状态修复或回调重放，应升级财务对账。

## 3. 版本控制

【PI-KB-003-6】订单更新使用乐观锁版本号。补偿动作必须携带诊断时读取的订单版本；版本变化时重新读取订单与支付记录，防止旧诊断覆盖新状态。
