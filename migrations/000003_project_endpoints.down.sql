DROP INDEX IF EXISTS mockmocks_project_method_path_idx;
DROP INDEX IF EXISTS mockmocks_global_method_path_idx;
ALTER TABLE mockmocks ADD CONSTRAINT mockmocks_method_path_key UNIQUE (method, path);
ALTER TABLE mockmocks DROP COLUMN IF EXISTS project_id;
