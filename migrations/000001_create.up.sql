CREATE TABLE teams (
  name text PRIMARY KEY NOT NULL
);

CREATE TABLE users (
  id serial PRIMARY KEY,
  user_id VARCHAR(4) NOT NULL,
  team_name text REFERENCES teams(name) ON DELETE CASCADE,
  name text NOT NULL,
  is_active boolean NOT NULL DEFAULT true
);

CREATE TYPE pr_status AS ENUM ('OPEN','MERGED');

CREATE TABLE prs (
  id text PRIMARY KEY,
  name text NOT NULL,
  author_id int REFERENCES users(id) NOT NULL,
  status pr_status NOT NULL DEFAULT 'OPEN',
  need_more_reviewers boolean NOT NULL DEFAULT false,
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now()
);

CREATE TABLE pr_reviewers (
  pr_id text REFERENCES prs(id) ON DELETE CASCADE,
  user_id int REFERENCES users(id),
  team_name text REFERENCES teams(name),
  assigned_at timestamptz DEFAULT now(),
  PRIMARY KEY (pr_id, user_id)
);