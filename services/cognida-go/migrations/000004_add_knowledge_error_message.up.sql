-- 文档处理失败时保存错误原因，确保前端状态可从 processing 正确切换为 failed。
ALTER TABLE `knowledge`
  ADD COLUMN `error_message` text COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `parse_status`;
