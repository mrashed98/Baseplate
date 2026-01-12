-- Super Admin Role Feature Migration
-- Extends users table with super admin designation
-- Extends audit_logs table with detailed action tracking

-- Add super admin fields to users table
ALTER TABLE users ADD COLUMN is_super_admin BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN super_admin_promoted_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN super_admin_promoted_by UUID REFERENCES users(id) ON DELETE SET NULL;

-- Add partial index for efficient super admin lookups
CREATE INDEX idx_users_super_admin ON users(is_super_admin)
  WHERE is_super_admin = true;

-- Extend audit_logs table for detailed super admin auditing
ALTER TABLE audit_logs ADD COLUMN actor_type VARCHAR(20) NOT NULL DEFAULT 'team_member'
  CHECK (actor_type IN ('team_member', 'super_admin', 'api_key'));
ALTER TABLE audit_logs ADD COLUMN ip_address INET;
ALTER TABLE audit_logs ADD COLUMN user_agent TEXT;
ALTER TABLE audit_logs ADD COLUMN result_status VARCHAR(20)
  CHECK (result_status IN ('success', 'failure', 'partial'));
ALTER TABLE audit_logs ADD COLUMN request_context JSONB NOT NULL DEFAULT '{}';

-- Add partial index for efficient super admin audit queries
CREATE INDEX idx_audit_logs_actor_type ON audit_logs(actor_type)
  WHERE actor_type = 'super_admin';

-- Note: Initial super admin creation handled by /cmd/init-superadmin tool
-- Requires environment variables: SUPER_ADMIN_EMAIL, SUPER_ADMIN_PASSWORD
