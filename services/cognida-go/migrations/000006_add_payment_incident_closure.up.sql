-- Case-closure metadata is descriptive and tenant scoped; no payment-operation data is stored here.
ALTER TABLE `payment_incidents`
  ADD COLUMN `recommendation_version` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' AFTER `assessment_request_id`,
  ADD COLUMN `evidence_snapshot` longtext COLLATE utf8mb4_unicode_ci NOT NULL AFTER `recommendation_version`,
  ADD COLUMN `conclusion` text COLLATE utf8mb4_unicode_ci NOT NULL AFTER `evidence_snapshot`;

CREATE TABLE `payment_incident_events` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `incident_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `tenant_id` bigint NOT NULL,
  `actor_id` bigint NOT NULL,
  `event_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `from_status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `to_status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `content` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `attachment_ref` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_payment_incident_events_tenant_incident_created` (`tenant_id`, `incident_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
