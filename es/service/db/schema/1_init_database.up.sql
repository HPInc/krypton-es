-- Initial schema for enroll database.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- verification public keys
CREATE TABLE public_key
(
	kid VARCHAR(64) NOT NULL,
	alg VARCHAR(16) NOT NULL,
	public_key TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	PRIMARY KEY(kid)
);
-- enroll data
CREATE TABLE enroll
(
	id UUID NOT NULL DEFAULT uuid_generate_v4(),
	request_id UUID NOT NULL DEFAULT uuid_generate_v4(),
	tenant_id TEXT NOT NULL,
	csr_hash TEXT NOT NULL,
	status SMALLINT NOT NULL DEFAULT 0,
	device_id UUID NULL,
	certificate TEXT NULL,
	parent_certificates TEXT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP NULL,
	PRIMARY KEY(id)
);
-- enroll_archive data
CREATE TABLE enroll_archive
(
	id UUID NOT NULL DEFAULT uuid_generate_v4(),
	request_id UUID NOT NULL DEFAULT uuid_generate_v4(),
	tenant_id TEXT NOT NULL,
	csr_hash TEXT NOT NULL,
	status SMALLINT NOT NULL DEFAULT 0,
	device_id UUID NULL,
	certificate TEXT NULL,
	parent_certificates TEXT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP NULL,
	PRIMARY KEY(id)
);
-- enroll_error data
CREATE TABLE enroll_error
(
	id UUID NOT NULL DEFAULT uuid_generate_v4(),
	request_id UUID NOT NULL DEFAULT uuid_generate_v4(),
	tenant_id TEXT NOT NULL,
	csr_hash TEXT NOT NULL,
	status SMALLINT NOT NULL DEFAULT 0,
	device_id UUID NULL,
	certificate TEXT NULL,
	error_code INTEGER NOT NULL DEFAULT 0,
	error_text TEXT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP NULL,
	PRIMARY KEY(id)
);
-- unenroll data
CREATE TABLE unenroll
(
	id UUID NOT NULL DEFAULT uuid_generate_v4(),
	request_id UUID NOT NULL DEFAULT uuid_generate_v4(),
	tenant_id TEXT NOT NULL,
	device_id UUID NOT NULL,
	status SMALLINT NOT NULL DEFAULT 0,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP NULL,
	PRIMARY KEY(id)
);
-- unenroll_archive data
CREATE TABLE unenroll_archive
(
	id UUID NOT NULL DEFAULT uuid_generate_v4(),
	request_id UUID NOT NULL DEFAULT uuid_generate_v4(),
	tenant_id TEXT NOT NULL,
	device_id UUID NOT NULL,
	status SMALLINT NOT NULL DEFAULT 0,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP NULL,
	PRIMARY KEY(id)
);
-- unenroll_error data
CREATE TABLE unenroll_error
(
	id UUID NOT NULL DEFAULT uuid_generate_v4(),
	request_id UUID NOT NULL,
	tenant_id TEXT NOT NULL,
	device_id UUID NOT NULL,
	status SMALLINT NOT NULL DEFAULT 0,
	error_code INTEGER NOT NULL DEFAULT 0,
	error_text TEXT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP NULL,
	PRIMARY KEY(id)
);
