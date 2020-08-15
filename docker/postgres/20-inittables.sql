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
