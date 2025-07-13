-- add user_id to all related enroll tables
-- enroll
ALTER TABLE enroll ADD user_id TEXT;
-- enroll_archive
ALTER TABLE enroll_archive ADD user_id TEXT;
-- enroll error
ALTER TABLE enroll_error ADD user_id TEXT;
--
-- unenroll
ALTER TABLE unenroll ADD user_id TEXT;
-- unenroll_archive
ALTER TABLE unenroll_archive ADD user_id TEXT;
-- unenroll error
ALTER TABLE unenroll_error ADD user_id TEXT;
--
