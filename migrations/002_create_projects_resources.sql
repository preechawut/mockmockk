CREATE TABLE projects (
 id uuid PRIMARY KEY, slug text NOT NULL UNIQUE, name text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE resources (
 id uuid PRIMARY KEY, project_id uuid NOT NULL REFERENCES projects(id),
 slug text NOT NULL, fields jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(project_id, slug), UNIQUE(project_id, id)
);
CREATE TABLE records (
 id uuid PRIMARY KEY, project_id uuid NOT NULL, resource_id uuid NOT NULL,
 data jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY(project_id, resource_id) REFERENCES resources(project_id, id)
);
CREATE INDEX records_resource_order ON records(resource_id, created_at, id);
