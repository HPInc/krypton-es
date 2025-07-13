-- add user_id to all related enroll tables
-- enroll
ALTER TABLE enroll DROP user_id;
-- enroll_archive
ALTER TABLE enroll_archive DROP user_id;
-- enroll error
ALTER TABLE enroll_error DROP user_id;
--
-- unenroll
ALTER TABLE unenroll DROP user_id;
-- unenroll_archive
ALTER TABLE unenroll_archive DROP user_id;
-- unenroll error
ALTER TABLE unenroll_error DROP user_id;
--
