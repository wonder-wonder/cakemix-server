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
-- CREATE TABLE IF NOT EXISTS project(
--   uuid TEXT PRIMARY KEY,
--   title TEXT,
--   description TEXT,
--   thumburi TEXT,
--   createat INTEGER,
--   isrecruting INTEGER,
--   categoryid INTEGER
-- );
-- CREATE TABLE IF NOT EXISTS projectmember(
--   uuid TEXT,
--   memberuuid TEXT,
--   permission INTEGER,
--   PRIMARY KEY (uuid, memberuuid),
--   FOREIGN KEY (uuid) REFERENCES project(uuid),
--   FOREIGN KEY (memberuuid) REFERENCES username(uuid)
-- );
-- CREATE TABLE IF NOT EXISTS projectpage(
--   uuid TEXT,
--   pagename TEXT,
--   title TEXT,
--   isdraft INTEGER,
--   isprivate INTEGER,
--   createat INTEGER,
--   createruuid TEXT,
--   updateat INTEGER,
--   updateruuid TEXT,
--   content TEXT,
--   PRIMARY KEY (uuid, pagename),
--   FOREIGN KEY (uuid) REFERENCES project(uuid),
--   FOREIGN KEY (createruuid) REFERENCES username(uuid),
--   FOREIGN KEY (updateruuid) REFERENCES username(uuid)
-- );
-- CREATE TABLE IF NOT EXISTS comment(
--   uuid TEXT, --ProjectUUID
--   cid TEXT PRIMARY KEY,
--   parentcid TEXT,
--   createat INTEGER,
--   createruuid TEXT,
--   content TEXT,
--   FOREIGN KEY (uuid) REFERENCES project(uuid),
--   FOREIGN KEY (createruuid) REFERENCES username(uuid)
-- );
-- CREATE TABLE IF NOT EXISTS projectfollow(
--   uuid TEXT,
--   followuuid TEXT,
--   followat INTEGER,
--   PRIMARY KEY (uuid, followuuid),
--   FOREIGN KEY (uuid) REFERENCES project(uuid),
--   FOREIGN KEY (followuuid) REFERENCES username(uuid)
-- );
-- CREATE TABLE IF NOT EXISTS projectgood(
--   uuid TEXT,
--   useruuid TEXT,
--   PRIMARY KEY (uuid, useruuid),
--   FOREIGN KEY (uuid) REFERENCES project(uuid),
--   FOREIGN KEY (useruuid) REFERENCES auth(uuid)
-- );
CREATE TABLE IF NOT EXISTS teammember(
  teamuuid TEXT,
  useruuid TEXT,
  permission INTEGER,
  joinat INTEGER,
  PRIMARY KEY (teamuuid, useruuid),
  FOREIGN KEY (teamuuid) REFERENCES username(UUID),
  FOREIGN KEY (useruuid) REFERENCES auth(uuid)
);
-- CREATE TABLE IF NOT EXISTS category(
--   categoryid SERIAL PRIMARY KEY,
--   name TEXT UNIQUE,
--   imgurl TEXT
-- );
-- CREATE TABLE IF NOT EXISTS tag(
--   tagid SERIAL PRIMARY KEY,
--   name TEXT UNIQUE
-- );
-- CREATE TABLE IF NOT EXISTS projecttag(
--   uuid TEXT,
--   tagid INTEGER,
--   PRIMARY KEY (uuid, tagid),
--   FOREIGN KEY (uuid) REFERENCES project(uuid),
--   FOREIGN KEY (tagid) REFERENCES tag(tagid)
-- );
