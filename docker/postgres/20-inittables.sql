\c cakemix;
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
  logindate INTEGER,
  lastdate INTEGER,
  expiredate INTEGER,
  ipaddr TEXT,
  devicedata TEXT,
  PRIMARY KEY (uuid, sessionid),
  FOREIGN KEY (uuid) REFERENCES auth(uuid)
);
-- User may fail so that uuid, username, and email can be duplicate. (System checks them when inserting)
CREATE TABLE IF NOT EXISTS preuser(
  uuid TEXT,
  username TEXT,
  email TEXT,
  password TEXT,
  salt TEXT,
  token TEXT PRIMARY KEY,
  expdate INTEGER
);
CREATE TABLE IF NOT EXISTS passreset(
  uuid TEXT,
  token TEXT PRIMARY KEY,
  expdate INTEGER,
  FOREIGN KEY (uuid) REFERENCES auth(uuid)
);
CREATE TABLE IF NOT EXISTS profile(
  uuid TEXT PRIMARY KEY,
  name TEXT,
  bio TEXT,
  iconuri TEXT,
  createat INTEGER,
  attr TEXT,
  lang TEXT,
  FOREIGN KEY (uuid) REFERENCES username(uuid)
);
CREATE TABLE IF NOT EXISTS teammember(
  teamuuid TEXT,
  useruuid TEXT,
  permission INTEGER,
  joinat INTEGER,
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
  createdat INTEGER,
  updatedat INTEGER,
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
  createdat INTEGER,
  updatedat INTEGER,
  updateruuid TEXT,
  tagid INTEGER,
  FOREIGN KEY (owneruuid) REFERENCES username(uuid),
  FOREIGN KEY (updateruuid) REFERENCES username(uuid),
  FOREIGN KEY (tagid) REFERENCES tag(tagid)
);
CREATE TABLE IF NOT EXISTS documentrevision(
  uuid TEXT PRIMARY KEY,
  text TEXT,
  updatedat INTEGER,
  FOREIGN KEY (uuid) REFERENCES document(uuid)
);