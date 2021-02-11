\c cakemix;
BEGIN;
CREATE TABLE IF NOT EXISTS username(uuid TEXT PRIMARY KEY, username TEXT UNIQUE);
CREATE TABLE IF NOT EXISTS auth(
  uuid TEXT PRIMARY KEY,
  email TEXT UNIQUE,
  password TEXT,
  salt TEXT,
  FOREIGN KEY (uuid) REFERENCES username(uuid)
);
CREATE TABLE IF NOT EXISTS session(
  uuid TEXT,
  sessionid TEXT,
  logindate BIGINT,
  lastdate BIGINT,
  expiredate BIGINT,
  ipaddr TEXT,
  devicedata TEXT,
  PRIMARY KEY (uuid, sessionid),
  FOREIGN KEY (uuid) REFERENCES auth(uuid)
);
CREATE TABLE IF NOT EXISTS invitetoken(
  fromuuid TEXT,
  token TEXT PRIMARY KEY,
  expdate BIGINT
);
-- User may fail so that uuid, username, and email can be duplicate. (System checks them when inserting)
CREATE TABLE IF NOT EXISTS preuser(
  uuid TEXT,
  username TEXT,
  email TEXT,
  password TEXT,
  salt TEXT,
  token TEXT PRIMARY KEY,
  expdate BIGINT
);
CREATE TABLE IF NOT EXISTS passreset(
  uuid TEXT,
  token TEXT PRIMARY KEY,
  expdate BIGINT,
  FOREIGN KEY (uuid) REFERENCES auth(uuid)
);
CREATE TABLE IF NOT EXISTS profile(
  uuid TEXT PRIMARY KEY,
  bio TEXT,
  iconuri TEXT,
  createat BIGINT,
  attr TEXT,
  lang TEXT,
  FOREIGN KEY (uuid) REFERENCES username(uuid)
);
CREATE TABLE IF NOT EXISTS teammember(
  teamuuid TEXT,
  useruuid TEXT,
  permission INTEGER,
  joinat BIGINT,
  PRIMARY KEY (teamuuid, useruuid),
  FOREIGN KEY (teamuuid) REFERENCES username(UUID),
  FOREIGN KEY (useruuid) REFERENCES auth(uuid)
);
CREATE TABLE IF NOT EXISTS tag(tagid SERIAL PRIMARY KEY, name TEXT UNIQUE);
CREATE TABLE IF NOT EXISTS folder(
  uuid TEXT PRIMARY KEY,
  owneruuid TEXT,
  parentfolderuuid TEXT,
  name TEXT,
  permission INTEGER,
  createdat BIGINT,
  updatedat BIGINT,
  updateruuid TEXT,
  FOREIGN KEY (owneruuid) REFERENCES username(uuid),
  FOREIGN KEY (updateruuid) REFERENCES username(uuid)
);
CREATE TABLE IF NOT EXISTS document(
  uuid TEXT PRIMARY KEY,
  owneruuid TEXT,
  parentfolderuuid TEXT,
  title TEXT,
  permission INTEGER,
  createdat BIGINT,
  updatedat BIGINT,
  updateruuid TEXT,
  tagid INTEGER,
  revision INTEGER,
  FOREIGN KEY (owneruuid) REFERENCES username(uuid),
  FOREIGN KEY (updateruuid) REFERENCES username(uuid),
  FOREIGN KEY (tagid) REFERENCES tag(tagid)
);
CREATE TABLE IF NOT EXISTS documentrevision(
  uuid TEXT,
  text TEXT,
  updatedat BIGINT,
  revision INTEGER,
  PRIMARY KEY (uuid, revision),
  FOREIGN KEY (uuid) REFERENCES document(uuid)
);
ALTER TABLE document ADD FOREIGN KEY (uuid,revision) REFERENCES documentrevision(uuid,revision);
CREATE TABLE IF NOT EXISTS log(
  uuid TEXT,
  date BIGINT,
  type TEXT,
  sessionid TEXT,
  targetuuid TEXT,
  targetfdid TEXT,
  extdataid BIGINT
);
CREATE TABLE IF NOT EXISTS logextloginpassreset(
  id BIGSERIAL PRIMARY KEY,
  ipaddr TEXT,
  devicedata TEXT
);
COMMIT;
