-- +goose Up
-- jwt_signing_keys table: Spec 004 security/auth data model (004-T008), per
-- specs/004-security-auth-model/data-model.md's "Migration Strategy" section.
CREATE TABLE jwt_signing_keys (
    key_id VARCHAR(64) PRIMARY KEY,
    public_key TEXT NOT NULL,
    private_key_secret_arn VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'retired')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    retire_at TIMESTAMP
);

-- +goose Down
DROP TABLE jwt_signing_keys;
