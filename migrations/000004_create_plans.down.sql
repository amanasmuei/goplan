-- Drop plans tables and related objects
DROP TRIGGER IF EXISTS update_milestones_updated_at ON milestones;
DROP TRIGGER IF EXISTS update_phases_updated_at ON phases;
DROP TRIGGER IF EXISTS update_plans_updated_at ON plans;
DROP TABLE IF EXISTS milestones CASCADE;
DROP TABLE IF EXISTS phases CASCADE;
DROP TABLE IF EXISTS plans CASCADE;
