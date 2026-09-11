ALTER TABLE mockmocks ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE CASCADE;

ALTER TABLE mockmocks DROP CONSTRAINT IF EXISTS mockmocks_method_path_key;

CREATE UNIQUE INDEX IF NOT EXISTS mockmocks_global_method_path_idx ON mockmocks(method, path) WHERE project_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS mockmocks_project_method_path_idx ON mockmocks(project_id, method, path) WHERE project_id IS NOT NULL;
