-- 支付故障案件：仅保存人工核验任务及其审计关联，不保存可执行的资金操作参数。
CREATE TABLE `payment_incidents` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `tenant_id` bigint NOT NULL,
  `created_by` bigint NOT NULL,
  `order_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `channel` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `incident_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `priority` varchar(8) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'P2',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending_verification',
  `assignee_id` bigint DEFAULT NULL,
  `assessment_request_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_payment_incidents_tenant_created` (`tenant_id`,`created_at`),
  KEY `idx_payment_incidents_tenant_status` (`tenant_id`,`status`),
  KEY `idx_payment_incidents_assessment_request` (`assessment_request_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
