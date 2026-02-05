-- Note: Dropping extensions may affect other databases using them
-- Only drop if you're certain no other schemas depend on these
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";
