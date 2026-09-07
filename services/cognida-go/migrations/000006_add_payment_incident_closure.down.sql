DROP TABLE `payment_incident_events`;
ALTER TABLE `payment_incidents`
  DROP COLUMN `conclusion`,
  DROP COLUMN `evidence_snapshot`,
  DROP COLUMN `recommendation_version`;
