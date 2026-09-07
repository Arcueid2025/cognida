-- Local, read-only fixtures for payment-incident verification. These tables
-- are not synchronized with any production payment, order or callback system.
CREATE TABLE `simulated_payment_orders` (
  `id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `tenant_id` bigint NOT NULL,
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `amount_minor` bigint NOT NULL,
  `currency` varchar(8) COLLATE utf8mb4_unicode_ci NOT NULL,
  `payment_channel` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_simulated_orders_tenant` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `simulated_payment_records` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `tenant_id` bigint NOT NULL,
  `order_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `provider_trade_no` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `amount_minor` bigint NOT NULL,
  `occurred_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_simulated_payments_tenant_order` (`tenant_id`, `order_id`, `occurred_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `simulated_payment_callback_logs` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `tenant_id` bigint NOT NULL,
  `order_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `payment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `event_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `delivery_status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `payload_summary` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `received_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_simulated_callbacks_tenant_order` (`tenant_id`, `order_id`, `received_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `simulated_payment_orders` (`id`, `tenant_id`, `status`, `amount_minor`, `currency`, `payment_channel`, `updated_at`) VALUES
  ('ORD-1020', 1, 'PAYING', 19900, 'CNY', 'wechat_pay', '2026-09-06 08:00:00.000'),
  ('ORD-SIM-PAID-001', 1, 'PAID', 9900, 'CNY', 'alipay', '2026-09-06 08:01:00.000');
INSERT INTO `simulated_payment_records` (`id`, `tenant_id`, `order_id`, `provider_trade_no`, `status`, `amount_minor`, `occurred_at`) VALUES
  ('SIM-PAY-1020-1', 1, 'ORD-1020', 'WX-SIM-1020', 'SUCCESS', 19900, '2026-09-06 07:59:56.000'),
  ('SIM-PAY-PAID-1', 1, 'ORD-SIM-PAID-001', 'ALI-SIM-001', 'SUCCESS', 9900, '2026-09-06 08:00:55.000');
INSERT INTO `simulated_payment_callback_logs` (`id`, `tenant_id`, `order_id`, `payment_id`, `event_type`, `delivery_status`, `payload_summary`, `received_at`) VALUES
  ('SIM-CB-1020-1', 1, 'ORD-1020', 'SIM-PAY-1020-1', 'payment.success', 'DELIVERY_FAILED', '模拟回调：上游支付成功；本地订单更新因 db_timeout 未完成。', '2026-09-06 08:00:01.000'),
  ('SIM-CB-PAID-1', 1, 'ORD-SIM-PAID-001', 'SIM-PAY-PAID-1', 'payment.success', 'DELIVERED', '模拟回调：支付成功并已更新订单状态。', '2026-09-06 08:01:01.000');
