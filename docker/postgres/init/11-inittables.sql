\c cakemix;
BEGIN;
CREATE TABLE IF NOT EXISTS username(uuid TEXT PRIMARY KEY, username TEXT UNIQUE NOT NULL);
CREATE TABLE IF NOT EXISTS auth(
  uuid TEXT PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  password TEXT NOT NULL,
  salt TEXT NOT NULL,
  FOREIGN KEY (uuid) REFERENCES username(uuid)
);
CREATE TABLE IF NOT EXISTS session(
  uuid TEXT NOT NULL,
  sessionid TEXT NOT NULL,
  logindate BIGINT NOT NULL,
  lastdate BIGINT NOT NULL,
  expiredate BIGINT NOT NULL,
  ipaddr TEXT NOT NULL,
  devicedata TEXT NOT NULL,
  PRIMARY KEY (uuid, sessionid),
  FOREIGN KEY (uuid) REFERENCES auth(uuid)
);
CREATE TABLE IF NOT EXISTS invitetoken(
  fromuuid TEXT NOT NULL,
  token TEXT PRIMARY KEY,
  expdate BIGINT NOT NULL
);
-- User may fail so that uuid, username, and email can be duplicate. (System checks them when inserting)
CREATE TABLE IF NOT EXISTS preuser(
  uuid TEXT NOT NULL,
  username TEXT NOT NULL,
  email TEXT NOT NULL,
  password TEXT NOT NULL,
  salt TEXT NOT NULL,
  token TEXT PRIMARY KEY,
  expdate BIGINT NOT NULL
);
CREATE TABLE IF NOT EXISTS passreset(
  uuid TEXT NOT NULL,
  token TEXT PRIMARY KEY,
  expdate BIGINT NOT NULL,
  FOREIGN KEY (uuid) REFERENCES auth(uuid)
);
CREATE TABLE IF NOT EXISTS profile(
  uuid TEXT PRIMARY KEY,
  bio TEXT NOT NULL,
  iconuri TEXT NOT NULL,
  createat BIGINT NOT NULL,
  attr TEXT NOT NULL,
  lang TEXT NOT NULL,
  FOREIGN KEY (uuid) REFERENCES username(uuid)
);
CREATE TABLE IF NOT EXISTS teammember(
  teamuuid TEXT NOT NULL,
  useruuid TEXT NOT NULL,
  permission INTEGER NOT NULL,
  joinat BIGINT NOT NULL,
  PRIMARY KEY (teamuuid, useruuid),
  FOREIGN KEY (teamuuid) REFERENCES username(UUID),
  FOREIGN KEY (useruuid) REFERENCES auth(uuid)
);
CREATE TABLE IF NOT EXISTS tag(tagid SERIAL PRIMARY KEY, name TEXT UNIQUE NOT NULL);
CREATE TABLE IF NOT EXISTS folder(
  uuid TEXT PRIMARY KEY,
  owneruuid TEXT NOT NULL,
  parentfolderuuid TEXT NOT NULL,
  name TEXT NOT NULL,
  permission INTEGER NOT NULL,
  createdat BIGINT NOT NULL,
  updatedat BIGINT NOT NULL,
  updateruuid TEXT NOT NULL,
  FOREIGN KEY (owneruuid) REFERENCES username(uuid),
  FOREIGN KEY (updateruuid) REFERENCES username(uuid)
);
CREATE TABLE IF NOT EXISTS document(
  uuid TEXT PRIMARY KEY,
  owneruuid TEXT NOT NULL,
  parentfolderuuid TEXT NOT NULL,
  title TEXT NOT NULL,
  permission INTEGER NOT NULL,
  createdat BIGINT NOT NULL,
  updatedat BIGINT NOT NULL,
  updateruuid TEXT NOT NULL,
  tagid INTEGER NOT NULL,
  revision INTEGER NOT NULL,
  FOREIGN KEY (owneruuid) REFERENCES username(uuid),
  FOREIGN KEY (updateruuid) REFERENCES username(uuid),
  FOREIGN KEY (tagid) REFERENCES tag(tagid)
);
CREATE TABLE IF NOT EXISTS documentrevision(
  uuid TEXT NOT NULL,
  text TEXT NOT NULL,
  updatedat BIGINT NOT NULL,
  revision INTEGER NOT NULL,
  PRIMARY KEY (uuid, revision),
  FOREIGN KEY (uuid) REFERENCES document(uuid)
);
CREATE TABLE IF NOT EXISTS log(
  uuid TEXT NOT NULL,
  date BIGINT NOT NULL,
  type TEXT NOT NULL,
  ipaddr TEXT NOT NULL,
  sessionid TEXT NOT NULL,
  targetuuid TEXT NOT NULL,
  targetfdid TEXT NOT NULL,
  extdataid BIGINT NOT NULL
);
CREATE TABLE IF NOT EXISTS logextloginpassreset(
  id BIGSERIAL PRIMARY KEY,
  devicedata TEXT NOT NULL
);
COMMIT;
