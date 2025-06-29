-- enroll policy
CREATE TABLE policy
(
	id UUID NOT NULL DEFAULT uuid_generate_v4(),
	tenant_id TEXT NOT NULL,
	data JSON NOT NULL,
	enabled BOOLEAN NOT NULL DEFAULT false,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP NULL,
	PRIMARY KEY(id)
);
create index policy_tenant_id_index on policy (tenant_id);
